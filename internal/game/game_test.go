package game

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestRulesAndScoring(t *testing.T) {
	s := New(TutorialPuzzle(), Standard, false, "fixture", time.Now())
	if s.Select(7, time.Now()) || len(s.Steps) != 0 {
		t.Fatal("first row restriction")
	}
	for i, cell := range s.Puzzle.Solution {
		if i == 1 && s.Select(1, time.Now()) {
			t.Fatal("reused cell")
		}
		if i == len(s.Puzzle.Solution)-1 {
			s.ElapsedMS = 24000
		}
		if !s.Select(cell, time.Now()) {
			t.Fatalf("step %d", i)
		}
		if i == 2 {
			if s.Uploaded() != 1 || !s.Legal(28) || s.Legal(27) {
				t.Fatal("alternating column or upload")
			}
			ok, k := Progress(s.Buffer(), s.Puzzle.Targets[1])
			if ok || k != 1 {
				t.Fatal("overlap progress")
			}
		}
	}
	if s.Reason != Complete || s.Score().Total != 2790 {
		t.Fatalf("result: %+v %+v", s, s.Score())
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	s = New(TutorialPuzzle(), Standard, false, "partial", time.Now())
	for _, cell := range s.Puzzle.Solution[:5] {
		s.Select(cell, time.Now())
	}
	s.End(Manual, time.Now())
	if s.Score().Total != 1125 {
		t.Fatal(s.Score())
	}
}
func TestMatchingSuffixAndPermanentUpload(t *testing.T) {
	ok, k := Progress([]string{"1C", "1C", "1C"}, []string{"1C", "1C", "55"})
	if ok || k != 2 {
		t.Fatal(ok, k)
	}
	ok, _ = Progress([]string{"1C", "55", "BD", "FF"}, []string{"1C", "55", "BD"})
	if !ok {
		t.Fatal("upload lost")
	}
}
func TestEndOrderingAndClock(t *testing.T) {
	s := New(TutorialPuzzle(), Standard, false, "last", time.Now())
	s.Puzzle.Board[0] = "FF"
	path := []int{0, 30, 31, 13, 16, 28, 26, 8}
	// A custom valid target completes on the final buffer cell.
	s.Puzzle.Targets = [][]string{{"1C", "7A", "55"}, {"55", "BD", "E9"}, {"E9", "7A", "55"}}
	for _, cell := range path {
		if !s.Select(cell, time.Now()) {
			t.Fatalf("invalid fixture %d", cell)
		}
	}
	if s.Reason != Complete {
		t.Fatal("capacity ended before upload", s.Reason)
	}
	s = New(TutorialPuzzle(), Standard, false, "timeout", time.Now())
	s.Advance(99999, time.Now())
	if s.ElapsedMS != 0 {
		t.Fatal("observation charged")
	}
	s.Select(1, time.Now())
	s.Advance(60000, time.Now())
	if s.Reason != Timeout || s.Select(13, time.Now()) || len(s.Steps) != 1 {
		t.Fatal("late input accepted")
	}
	s = New(TutorialPuzzle(), Standard, false, "empty", time.Now())
	s.End(Manual, time.Now())
	if s.Score().Total != 0 {
		t.Fatal("empty score")
	}
}
func TestPracticeUndoAndValidation(t *testing.T) {
	s := New(TutorialPuzzle(), Practice, true, "practice", time.Now())
	s.Select(1, time.Now())
	s.Advance(90000, time.Now())
	if !s.Undo() || s.ElapsedMS != 90000 || s.Undos != 1 {
		t.Fatal(s)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	s.Mode = Standard
	if s.Validate() == nil || s.Undo() {
		t.Fatal("assisted standard accepted")
	}
}

func TestTutorialValidation(t *testing.T) {
	newTutorial := func() Session {
		return New(TutorialPuzzle(), Tutorial, true, "tutorial", time.Now())
	}
	s := newTutorial()
	for _, cell := range s.Puzzle.Solution[:3] {
		s.Select(cell, time.Now())
	}
	s.Resumes = 2
	if err := s.Validate(); err != nil {
		t.Fatalf("valid resumed tutorial rejected: %v", err)
	}
	for _, test := range []struct {
		name   string
		mutate func(*Session)
	}{
		{"timed", func(s *Session) { s.Unlimited = false }},
		{"hints", func(s *Session) { s.Hints = 1 }},
		{"undos", func(s *Session) { s.Undos = 1 }},
		{"wrong path", func(s *Session) { s.Select(0, time.Now()) }},
		{"past solution", func(s *Session) {
			for _, cell := range s.Puzzle.Solution {
				s.Select(cell, time.Now())
			}
			s.Steps = append(s.Steps, Step{Cell: 0})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := newTutorial()
			test.mutate(&s)
			if err := s.Validate(); err == nil {
				t.Fatal("invalid tutorial accepted")
			}
		})
	}
}
func TestGenerationProperties(t *testing.T) {
	for d := Difficulty(0); d < 4; d++ {
		for seed := uint64(0); seed < 250; seed++ {
			p, err := Generate(context.Background(), seed, d)
			if err != nil {
				t.Fatal(d, seed, err)
			}
			if err = p.Validate(); err != nil {
				t.Fatal(err)
			}
			r := d.Rules()
			if MinimumCodes(p.Targets) < r.MinPath || len(p.Solution) > r.MaxPath {
				t.Fatal("difficulty outside bounds")
			}
			again, _ := Generate(context.Background(), seed, d)
			if !reflect.DeepEqual(p, again) {
				t.Fatal("unstable generator")
			}
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Generate(ctx, 1, 0); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
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
