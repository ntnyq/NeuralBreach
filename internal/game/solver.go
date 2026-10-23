package game

import (
	"context"
	"errors"
	"slices"
)

var ErrBudget = errors.New("search budget exceeded")

var ErrNoSolution = errors.New("no full upload reachable; undo or restart")

// Solve returns a continuation, not a path from a different state. Search has
// both cancellation and a hard node budget; exhausting it is not 'no solution'.
func Solve(ctx context.Context, s Session, budget int) ([]int, error) {
	var answer []int
	var searchErr error
	var visit func(Session, []int) bool
	visit = func(state Session, path []int) bool {
		budget--
		if budget < 0 {
			searchErr = ErrBudget
			return false
		}
		if err := ctx.Err(); err != nil {
			searchErr = err
			return false
		}
		if state.Uploaded() == len(state.Puzzle.Targets) {
			answer = slices.Clone(path)
			return true
		}
		remaining := state.Puzzle.Difficulty.Rules().Capacity - len(state.Steps)
		if remaining == 0 {
			return false
		}
		b := state.Buffer()
		for _, target := range state.Puzzle.Targets {
			ok, k := Progress(b, target)
			if !ok && len(target)-k > remaining {
				return false
			}
		}
		for _, cell := range state.Moves() {
			next := state
			next.Steps = append(slices.Clone(state.Steps), Step{Cell: cell})
			if visit(next, append(path, cell)) {
				return true
			}
			if searchErr != nil {
				return false
			}
		}
		return false
	}
	if visit(s, nil) {
		return answer, nil
	}
	if searchErr != nil {
		return nil, searchErr
	}
	return nil, ErrNoSolution
}
