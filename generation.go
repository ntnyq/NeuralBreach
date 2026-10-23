package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/locales"
)

var errUniquePuzzle = errors.New("could not generate a new puzzle; please try again")

// generatePuzzle leaves seeded practice deterministic. Random games must reserve
// their content durably before returning, with bounded retries and cancellation.
func generatePuzzle(ctx context.Context, level game.Difficulty, seed *uint64, nextSeed func() (uint64, error), reserve func(context.Context, game.Puzzle) (bool, error)) (game.Puzzle, error) {
	if seed != nil {
		return game.Generate(ctx, *seed, level)
	}
	for range 128 {
		if err := ctx.Err(); err != nil {
			return game.Puzzle{}, err
		}
		next, err := nextSeed()
		if err != nil {
			return game.Puzzle{}, err
		}
		p, err := game.Generate(ctx, next, level)
		if err != nil {
			return game.Puzzle{}, err
		}
		accepted, err := reserve(ctx, p)
		if err != nil {
			return game.Puzzle{}, err
		}
		if accepted {
			return p, nil
		}
	}
	return game.Puzzle{}, errUniquePuzzle
}

func randomSeed() (uint64, error) {
	var b [8]byte
	_, err := rand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:]), err
}

func (w *window) generate() {
	var seed *uint64
	if w.mode == game.Practice && w.seed != "" {
		value, err := strconv.ParseUint(w.seed, 10, 64)
		if err != nil {
			w.message = locales.InvalidSeed
			return
		}
		seed = &value
	}
	id, err := randomSeed()
	if err != nil {
		w.errorDetail = err
		return
	}
	level, mode, unlimited := w.difficulty, w.mode, w.unlimited
	// Include unsaved sessions without reading mutable UI state in the worker.
	seen := maps.Clone(w.data.PuzzleFingerprints)
	store := w.store
	reserve := func(ctx context.Context, p game.Puzzle) (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if seen[p.Fingerprint()] {
			return false, nil
		}
		if store != nil {
			return store.ReservePuzzle(ctx, p)
		}
		return true, nil
	}
	var puzzle game.Puzzle
	w.message = locales.Generating
	w.run(func(ctx context.Context) error {
		var err error
		puzzle, err = generatePuzzle(ctx, level, seed, randomSeed, reserve)
		return err
	}, func() {
		s := game.New(puzzle, mode, unlimited, fmt.Sprintf("%x-%x", time.Now().UnixNano(), id), time.Now())
		w.data.RememberPuzzle(puzzle)
		w.session = &s
		w.focusCell = 0
		w.focusBoard = true
		w.hintCell = -1
		w.lastTick = time.Now()
		w.navigate(gamePage)
		w.persist(nil)
	})
}
