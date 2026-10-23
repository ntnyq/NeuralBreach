package records

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ntnyq/cyber-game/internal/game"
)

func finished(id string) game.Session {
	s := game.New(game.TutorialPuzzle(), game.Standard, false, id, time.Now())
	for _, cell := range s.Puzzle.Solution {
		s.Select(cell, time.Now())
	}
	return s
}
func TestCommitResumeAndIdempotentSettlement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, d, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s := game.New(game.TutorialPuzzle(), game.Standard, false, "run", time.Now())
	s.Select(1, time.Now())
	s.Advance(1234, time.Now())
	d.Active = &s
	d.Revision = 1
	if err = store.Commit(d); err != nil {
		t.Fatal(err)
	}
	_, restored, err := Open(path)
	if err != nil || restored.Active.ElapsedMS != 1234 {
		t.Fatal(restored, err)
	}
	s = finished("run")
	d.Finish(s)
	d.Finish(s)
	d.Revision = 2
	if err = store.Commit(d); err != nil {
		t.Fatal(err)
	}
	_, restored, err = Open(path)
	if err != nil || restored.Active != nil || len(restored.History) != 1 {
		t.Fatal(restored, err)
	}
}
func TestOlderWritesCannotWin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, d, _ := Open(path)
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		v := d.Clone()
		v.Revision = uint64(i)
		v.Settings.Master = float64(i)
		wg.Go(func() {
			if err := store.Commit(v); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	_, got, err := Open(path)
	if err != nil || got.Revision != 30 || got.Settings.Master != 30 {
		t.Fatal(got, err)
	}
}
func TestCorruptionProtectedAndBackup(t *testing.T) {
	for _, raw := range []string{
		"broken", "", "{}", "null", "[]", "true",
		`{"settings":{"scale":100,"language":"en"}}`,
		`{"version":null}`, `{"version":0}`, `{"version":-1}`,
		`{"version":"1"}`, `{"version":99}`, `{"version":1,"settings":{"scale":900}}`,
	} {
		t.Run(raw, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.json")
			if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
				t.Fatal(err)
			}
			store, d, err := Open(path)
			if err == nil {
				t.Fatal("accepted corrupt data")
			}
			d.Revision = 1
			if store.Commit(d) == nil {
				t.Fatal("overwrote corrupt data")
			}
			b, _ := os.ReadFile(path)
			if string(b) != raw {
				t.Fatal("changed bytes")
			}
			if _, err = store.Repair(); err != nil {
				t.Fatal(err)
			}
			backups, _ := filepath.Glob(path + ".backup-*")
			if len(backups) != 1 {
				t.Fatal(backups)
			}
			b, _ = os.ReadFile(backups[0])
			if string(b) != raw {
				t.Fatal("backup differs")
			}
			if _, _, err = Open(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNewStoreAndVersionedSettingsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, d, err := Open(path)
	if err != nil || d.Version != 1 || d.Settings != Defaults() {
		t.Fatalf("new store: %+v, %v", d, err)
	}
	d.Revision++
	if err := store.Commit(d); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"settings":{"muted":true}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, d, err = Open(path)
	want := Defaults()
	want.Muted = true
	if err != nil || d.Settings != want {
		t.Fatalf("versioned settings defaults: %+v, %v", d, err)
	}
}

func TestTutorialSaveAndSettlement(t *testing.T) {
	for _, reason := range []game.Reason{game.Complete, game.Abandoned} {
		t.Run(string(reason), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.json")
			store, d, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			s := game.New(game.TutorialPuzzle(), game.Tutorial, true, "tutorial", time.Now())
			s.Select(s.Puzzle.Solution[0], time.Now())
			s.Advance(1234, time.Now())
			d.Active = &s
			d.History = []game.Session{finished("previous")}
			d.Revision++
			if err := store.Commit(d); err != nil {
				t.Fatal(err)
			}
			_, restored, err := Open(path)
			if err != nil || restored.Active == nil {
				t.Fatalf("missing tutorial save: %v", err)
			}
			s = *restored.Active
			if s.Mode != game.Tutorial || !s.Unlimited || len(s.Steps) != 1 || s.ElapsedMS != 1234 {
				t.Fatalf("tutorial changed on reload: %+v", s)
			}
			if reason == game.Complete {
				for _, cell := range s.Puzzle.Solution[1:] {
					s.Select(cell, time.Now())
				}
			} else {
				s.End(reason, time.Now())
			}
			restored.Finish(s)
			restored.Finish(s)
			restored.Revision++
			if err := store.Commit(restored); err != nil {
				t.Fatal(err)
			}
			_, restored, err = Open(path)
			if err != nil || restored.Active != nil || len(restored.History) != 1 || restored.History[0].ID != "previous" {
				t.Fatalf("tutorial polluted history or left an active slot: %+v, %v", restored, err)
			}
			played, wins, _ := restored.Stats()
			if played != 1 || wins != 1 || restored.Best(1).ID != "previous" {
				t.Fatal("tutorial changed standard statistics")
			}
		})
	}
}
func TestWriteFailureAndBestExclusions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "obstacle")
	if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	store, d, _ := Open(filepath.Join(path, "data.json"))
	d.Revision = 1
	if store.Commit(d) == nil {
		t.Fatal("write should fail")
	}
	d = Empty()
	practice := finished("practice")
	practice.Mode = game.Practice
	abandoned := finished("abandoned")
	abandoned.Reason = game.Abandoned
	d.History = []game.Session{practice, abandoned}
	if d.Best(1) != nil {
		t.Fatal("practice or abandoned ranked")
	}
	standard := finished("standard")
	d.History = append(d.History, standard)
	if d.Best(1).ID != "standard" {
		t.Fatal("best missing")
	}
	played, wins, _ := d.Stats()
	if played != 1 || wins != 1 {
		t.Fatal("wrong stats")
	}
}
