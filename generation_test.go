package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/records"
)

func TestRandomGenerationSkipsPersistedContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, _, err := records.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	p, err := game.Generate(ctx, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	// Identity is content, even when the saved seed and witness differ.
	p.Seed = 999
	if ok, err := store.ReservePuzzle(ctx, p); !ok || err != nil {
		t.Fatal(ok, err)
	}
	store, _, err = records.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	var calls uint64
	got, err := generatePuzzle(ctx, 1, nil, func() (uint64, error) {
		calls++
		return calls, nil
	}, store.ReservePuzzle)
	if err != nil || calls != 2 || got.Seed != 2 || got.Fingerprint() == p.Fingerprint() {
		t.Fatalf("duplicate not retried: seed=%d, calls=%d, err=%v", got.Seed, calls, err)
	}
	_, d, err := records.Open(path)
	if err != nil || !d.PuzzleFingerprints[got.Fingerprint()] {
		t.Fatalf("returned before durable reservation: %v", err)
	}

	seed := uint64(1)
	want, _ := game.Generate(ctx, seed, 1)
	for range 2 {
		got, err := generatePuzzle(ctx, 1, &seed, func() (uint64, error) {
			t.Fatal("seeded practice requested a random seed")
			return 0, nil
		}, func(context.Context, game.Puzzle) (bool, error) {
			t.Fatal("seeded practice tried to reserve a duplicate")
			return false, nil
		})
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("seeded practice changed", err)
		}
	}
}

func TestRandomGenerationFailureAndCancellation(t *testing.T) {
	ctx := context.Background()
	seed := func() (uint64, error) { return 1, nil }
	failed := errors.New("failed")
	if _, err := generatePuzzle(ctx, 0, nil, func() (uint64, error) { return 0, failed }, nil); !errors.Is(err, failed) {
		t.Fatal(err)
	}
	if p, err := generatePuzzle(ctx, 0, nil, seed, func(context.Context, game.Puzzle) (bool, error) {
		return false, failed
	}); !errors.Is(err, failed) || p.Board != nil {
		t.Fatal("returned a puzzle after reservation failed", err)
	}
	calls := 0
	if _, err := generatePuzzle(ctx, 0, nil, seed, func(context.Context, game.Puzzle) (bool, error) {
		calls++
		return false, nil
	}); !errors.Is(err, errUniquePuzzle) || calls != 128 {
		t.Fatal("duplicate retries were not bounded", calls, err)
	}
	canceled, cancel := context.WithCancel(ctx)
	calls = 0
	if _, err := generatePuzzle(canceled, 0, nil, seed, func(context.Context, game.Puzzle) (bool, error) {
		calls++
		cancel()
		return false, nil
	}); !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal("retry ignored cancellation", calls, err)
	}
}

func TestWindowRandomAndSeededGeneration(t *testing.T) {
	for _, mode := range []game.Mode{game.Standard, game.Practice} {
		t.Run(string(mode), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.json")
			w, wait := persistentWindow(t, path)
			w.mode = mode
			w.navigate(setupPage)
			w.generate()
			if w.page != setupPage || w.session != nil {
				t.Fatal("random puzzle exposed before reservation completed")
			}
			wait()
			if w.session == nil || w.session.Mode != mode || w.page != gamePage {
				t.Fatalf("random game did not start: %v", w.errorDetail)
			}
			p := w.session.Puzzle
			_, d, err := records.Open(path)
			if err != nil || !d.PuzzleFingerprints[p.Fingerprint()] || d.Active == nil {
				t.Fatal("random game not durably indexed and saved", err)
			}
			for range 2 {
				w.abandon()
				w.mode = game.Practice
				w.seed = strconv.FormatUint(p.Seed, 10)
				w.generate()
				wait()
				if w.session == nil || w.session.Mode != game.Practice || !reflect.DeepEqual(w.session.Puzzle, p) {
					t.Fatalf("seeded practice rejected or changed an indexed puzzle: %v", w.errorDetail)
				}
			}
		})
	}
}

func TestWindowDoesNotExposeRandomPuzzleOnWriteFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	w, wait := persistentWindow(t, path)
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	w.navigate(setupPage)
	w.generate()
	wait()
	if w.errorDetail == nil || w.session != nil || w.page != setupPage || len(w.data.PuzzleFingerprints) != 0 {
		t.Fatal("failed reservation exposed or consumed a random puzzle")
	}
}

func TestCanceledGenerationKeepsDurableReservation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	store, data, err := records.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	w := newWindow(store, data, nil)
	t.Cleanup(w.close)
	queue := make(chan func(), 1)
	w.postUpdate = func(fn func()) { queue <- fn }
	w.navigate(setupPage)
	w.generate()
	select {
	case done := <-queue:
		w.navigate(menuPage)
		done()
	case <-time.After(5 * time.Second):
		t.Fatal("generation did not finish")
	}
	if w.session != nil || w.page != menuPage {
		t.Fatal("canceled completion opened the puzzle")
	}
	// Even a later settings save whose UI map never saw the reservation keeps it.
	w.data.Revision++
	if err := store.Commit(w.data.Clone()); err != nil {
		t.Fatal(err)
	}
	_, restored, err := records.Open(path)
	if err != nil || restored.Active != nil || len(restored.PuzzleFingerprints) != 1 {
		t.Fatal("canceled callback lost a durable reservation", err)
	}
}
