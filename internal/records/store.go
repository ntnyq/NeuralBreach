// Package records persists preferences, the active slot and history together.
package records

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/storage"
)

type Settings struct {
	Master         float64 `json:"master"`
	Music          float64 `json:"music"`
	Effects        float64 `json:"effects"`
	Muted          bool    `json:"muted"`
	Dynamic        bool    `json:"dynamic"`
	BackgroundMute bool    `json:"backgroundMute"`
	Glitch         int     `json:"glitch"`
	Reduced        bool    `json:"reducedMotion"`
	LowTimePulse   bool    `json:"lowTimePulse"`
	Scale          int     `json:"scale"`
	Language       string  `json:"language"`
}

func Defaults() Settings {
	return Settings{Master: 80, Music: 65, Effects: 85, Dynamic: true, BackgroundMute: true, Glitch: 1, Scale: 100, Language: "zh-Hans"}
}

type Data struct {
	Version            int             `json:"version"`
	Revision           uint64          `json:"revision"`
	Settings           Settings        `json:"settings"`
	Active             *game.Session   `json:"active,omitempty"`
	History            []game.Session  `json:"history"`
	PuzzleFingerprints map[string]bool `json:"puzzleFingerprints,omitempty"`
}

func Empty() Data { return Data{Version: 1, Settings: Defaults(), History: []game.Session{}} }

func (d Data) Clone() Data {
	b, _ := json.Marshal(d)
	var out Data
	_ = json.Unmarshal(b, &out)
	return out
}

func (d *Data) Finish(s game.Session) {
	if s.Mode != game.Tutorial {
		d.RememberPuzzle(s.Puzzle)
		found := false
		for _, old := range d.History {
			if old.ID == s.ID {
				found = true
				break
			}
		}
		if !found {
			d.History = append([]game.Session{s}, d.History...)
		}
	}
	if d.Active != nil && d.Active.ID == s.ID {
		d.Active = nil
	}
}

func (d *Data) RememberPuzzle(p game.Puzzle) {
	if d.PuzzleFingerprints == nil {
		d.PuzzleFingerprints = make(map[string]bool)
	}
	d.PuzzleFingerprints[p.Fingerprint()] = true
}

// indexPuzzles also upgrades snapshots written before fingerprints were added.
func (d *Data) indexPuzzles() {
	for _, s := range d.History {
		d.RememberPuzzle(s.Puzzle)
	}
	if d.Active != nil && d.Active.Mode != game.Tutorial {
		d.RememberPuzzle(d.Active.Puzzle)
	}
}

func (d Data) Validate() error {
	if d.Version != 1 {
		return errors.New("unsupported data version")
	}
	for fingerprint, seen := range d.PuzzleFingerprints {
		b, err := hex.DecodeString(fingerprint)
		if err != nil || len(b) != 32 || fingerprint != strings.ToLower(fingerprint) || !seen {
			return errors.New("invalid puzzle fingerprint")
		}
	}
	s := d.Settings
	if s.Master < 0 || s.Master > 100 || s.Music < 0 || s.Music > 100 || s.Effects < 0 || s.Effects > 100 || s.Glitch < 0 || s.Glitch > 3 || (s.Scale != 100 && s.Scale != 125 && s.Scale != 150) {
		return errors.New("invalid preferences")
	}
	switch s.Language {
	case "en", "zh-Hans", "zh-Hant", "ja", "ko":
	default:
		return errors.New("unsupported language")
	}
	ids := map[string]bool{}
	for _, record := range d.History {
		if record.Reason == "" || record.Mode == game.Tutorial || ids[record.ID] {
			return errors.New("invalid history")
		}
		if err := record.Validate(); err != nil {
			return err
		}
		ids[record.ID] = true
	}
	if d.Active != nil {
		if d.Active.Reason != "" || ids[d.Active.ID] {
			return errors.New("invalid active slot")
		}
		return d.Active.Validate()
	}
	return nil
}

type Store struct {
	mu       sync.Mutex
	path     string
	revision uint64
	blocked  error
	data     Data
}

func Open(path string) (*Store, Data, error) {
	s := &Store{path: path, data: Empty()}
	// Existing files must supply their version; defaults must not disguise
	// a missing version or a JSON null as a valid empty store.
	d := Data{Settings: Defaults(), History: []game.Session{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, Empty(), nil
	}
	if err == nil {
		err = json.Unmarshal(b, &d)
		if err == nil {
			err = d.Validate()
		}
	}
	if err != nil {
		s.blocked = err
		return s, Empty(), err
	}
	s.revision = d.Revision
	d.indexPuzzles()
	s.data = d.Clone()
	return s, d, nil
}

// Commit serializes writers. A late job can never overwrite a newer revision.
func (s *Store) Commit(d Data) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.blocked != nil {
		return fmt.Errorf("data protected: %w", s.blocked)
	}
	if d.Revision <= s.revision {
		return nil
	}
	if err := d.Validate(); err != nil {
		return err
	}
	d = d.Clone()
	d.indexPuzzles()
	// Reservations may finish after the UI captured this snapshot. They are
	// monotonic and must survive even a newer settings or session commit.
	for fingerprint := range s.data.PuzzleFingerprints {
		if d.PuzzleFingerprints == nil {
			d.PuzzleFingerprints = make(map[string]bool)
		}
		d.PuzzleFingerprints[fingerprint] = true
	}
	return s.write(d)
}

// ReservePuzzle records a random puzzle before it can be shown. Duplicate
// reservations return false; failed writes neither reserve nor expose a puzzle.
// Reservations share the snapshot but do not advance the UI-owned revision.
func (s *Store) ReservePuzzle(ctx context.Context, p game.Puzzle) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if s.blocked != nil {
		return false, fmt.Errorf("data protected: %w", s.blocked)
	}
	if err := p.Validate(); err != nil {
		return false, err
	}
	if s.data.PuzzleFingerprints[p.Fingerprint()] {
		return false, nil
	}
	d := s.data.Clone()
	d.RememberPuzzle(p)
	if err := s.write(d); err != nil {
		return false, err
	}
	return true, nil
}

// write requires mu and takes ownership of an isolated snapshot.
func (s *Store) write(d Data) error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	if err = storage.WriteFile(s.path, b, 0o600); err != nil {
		return err
	}
	s.revision = d.Revision
	s.data = d
	return nil
}

// Repair preserves the exact unreadable bytes before creating a fresh store.
func (s *Store) Repair() (Data, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.path)
	if err != nil {
		return Data{}, err
	}
	backup := s.path + ".backup-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if err = storage.WriteFile(backup, b, 0o600); err != nil {
		return Data{}, err
	}
	d := Empty()
	d.Revision = s.revision + 1
	b, err = json.MarshalIndent(d, "", "  ")
	if err != nil {
		return Data{}, err
	}
	if err = storage.WriteFile(s.path, b, 0o600); err != nil {
		return Data{}, err
	}
	s.blocked = nil
	s.revision = d.Revision
	s.data = d.Clone()
	return d, nil
}
