package records

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ntnyq/cyber-game/internal/game"
)

func generated(t *testing.T, seed uint64) game.Puzzle {
	t.Helper()
	p, err := game.Generate(context.Background(), seed, 1)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLegacyPuzzlesBackfillFingerprints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	d := Empty()
	for i, mode := range []game.Mode{game.Standard, game.Practice} {
		s := game.New(generated(t, uint64(i)), mode, false, string(mode), time.Now())
		s.End(game.Abandoned, time.Now())
		d.History = append(d.History, s)
	}
	s := game.New(generated(t, 2), game.Practice, false, "active", time.Now())
	d.Active = &s
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	store, restored, err := Open(path)
	if err != nil || len(restored.PuzzleFingerprints) != 3 {
		t.Fatalf("legacy puzzles not indexed: %+v, %v", restored.PuzzleFingerprints, err)
	}
	for _, p := range []game.Puzzle{d.History[0].Puzzle, d.History[1].Puzzle, s.Puzzle} {
		if ok, err := store.ReservePuzzle(context.Background(), p); ok || err != nil {
			t.Fatal("legacy content was accepted", ok, err)
		}
	}
	// An ordinary commit persists the backfill, independently of retained history.
	restored.History, restored.Active = nil, nil
	restored.Revision++
	if err := store.Commit(restored); err != nil {
		t.Fatal(err)
	}
	_, restored, err = Open(path)
	if err != nil || len(restored.PuzzleFingerprints) != 3 {
		t.Fatal("fingerprints lost on reload", err)
	}
}

func TestReservationsSurviveStaleSnapshotsAndRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, d, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	active := game.New(generated(t, 10), game.Standard, false, "active", time.Now())
	d.Active = &active
	d.Settings.Muted = true
	d.Revision = 1
	if err := store.Commit(d); err != nil {
		t.Fatal(err)
	}
	p := generated(t, 11)
	if ok, err := store.ReservePuzzle(context.Background(), p); !ok || err != nil {
		t.Fatal(ok, err)
	}
	_, reserved, err := Open(path)
	if err != nil || reserved.Revision != 1 || reserved.Active.ID != active.ID || !reserved.Settings.Muted {
		t.Fatal("reservation changed unrelated snapshot state", err)
	}
	// This UI snapshot predates the reservation, but has a newer revision.
	d.Active = nil
	d.Revision = 2
	if err := store.Commit(d); err != nil {
		t.Fatal(err)
	}
	d.Revision = 1
	if err := store.Commit(d); err != nil {
		t.Fatal(err)
	}
	store, restored, err := Open(path)
	if err != nil || restored.Active != nil || len(restored.PuzzleFingerprints) != 2 {
		t.Fatal("stale snapshot erased fingerprints or restored old active slot", err)
	}
	if ok, err := store.ReservePuzzle(context.Background(), p); ok || err != nil {
		t.Fatal("reserved puzzle repeated after restart", ok, err)
	}
	clone := restored.Clone()
	delete(clone.PuzzleFingerprints, p.Fingerprint())
	if !restored.PuzzleFingerprints[p.Fingerprint()] {
		t.Fatal("clone shares fingerprint map")
	}
}

func TestConcurrentReservationsAcceptOnce(t *testing.T) {
	store, _, err := Open(filepath.Join(t.TempDir(), "data.json"))
	if err != nil {
		t.Fatal(err)
	}
	p := generated(t, 1)
	results := make(chan bool, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			ok, err := store.ReservePuzzle(context.Background(), p)
			if err != nil {
				t.Error(err)
			}
			results <- ok
		})
	}
	wg.Wait()
	close(results)
	accepted := 0
	for ok := range results {
		if ok {
			accepted++
		}
	}
	if accepted != 1 {
		t.Fatal("accepted duplicate reservations", accepted)
	}
}

func TestReservationFailureDoesNotConsumePuzzle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, _, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	p := generated(t, 1)
	if ok, err := store.ReservePuzzle(context.Background(), p); ok || err == nil {
		t.Fatal("accepted puzzle despite failed write", ok, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if ok, err := store.ReservePuzzle(ctx, p); ok || !errors.Is(err, context.Canceled) {
		t.Fatal("canceled reservation accepted", ok, err)
	}
	if ok, err := store.ReservePuzzle(context.Background(), p); !ok || err != nil {
		t.Fatal("failed reservation consumed the puzzle", ok, err)
	}
}

func TestInvalidFingerprintsProtectStore(t *testing.T) {
	for _, fingerprints := range []map[string]bool{
		{"broken": true},
		{game.TutorialPuzzle().Fingerprint(): false},
	} {
		path := filepath.Join(t.TempDir(), "data.json")
		d := Empty()
		d.PuzzleFingerprints = fingerprints
		b, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b, 0o600); err != nil {
			t.Fatal(err)
		}
		store, _, err := Open(path)
		if err == nil {
			t.Fatal("accepted invalid fingerprint index")
		}
		if ok, err := store.ReservePuzzle(context.Background(), generated(t, 1)); ok || err == nil {
			t.Fatal("reservation bypassed corrupt store protection")
		}
		after, err := os.ReadFile(path)
		if err != nil || string(after) != string(b) {
			t.Fatal("changed protected bytes", err)
		}
	}
}
