package main

import (
	"context"
	"errors"

	"github.com/ntnyq/cyber-game/internal/game"
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
