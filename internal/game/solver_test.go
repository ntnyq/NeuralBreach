package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSolverContinuationAndBudget(t *testing.T) {
	s := New(TutorialPuzzle(), Practice, true, "hint", time.Now())
	for _, cell := range s.Puzzle.Solution[:3] {
		s.Select(cell, time.Now())
	}
	path, err := Solve(context.Background(), s, 250000)
	if err != nil {
		t.Fatal(err)
	}
	for _, cell := range path {
		if !s.Select(cell, time.Now()) {
			t.Fatal("illegal hint")
		}
	}
	if s.Reason != Complete {
		t.Fatal("hint not a full solution")
	}
	s = New(TutorialPuzzle(), Practice, true, "budget", time.Now())
	if _, err := Solve(context.Background(), s, 0); !errors.Is(err, ErrBudget) {
		t.Fatal(err)
	}
}
