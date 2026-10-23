package game

import (
	"context"
	"errors"
	"slices"
	"time"
)

// splitmix64 is fixed here instead of relying on the Go random implementation.
type random uint64

func (r *random) next(n int) int {
	*r += 0x9e3779b97f4a7c15
	z := uint64(*r)
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return int((z ^ (z >> 31)) % uint64(n))
}

// MinimumCodes is the shortest common superstring length, ignoring the board.
// With at most four non-contained targets, enumerating all permutations is exact.
// This is a lower bound on every legal board solution.
func MinimumCodes(targets [][]string) int {
	best := 100
	var visit func([]string, uint)
	visit = func(buffer []string, mask uint) {
		if len(buffer) >= best {
			return
		}
		if mask == (1<<len(targets))-1 {
			best = len(buffer)
			return
		}
		for i, target := range targets {
			if mask&(1<<i) != 0 {
				continue
			}
			if ok, _ := Progress(buffer, target); ok {
				visit(buffer, mask|1<<i)
				continue
			}
			_, k := Progress(buffer, target)
			next := append(slices.Clone(buffer), target[k:]...)
			visit(next, mask|1<<i)
		}
	}
	visit(nil, 0)
	return best
}

// Generate plants a joint solution and proves its length: the unconstrained
// lower bound and a legal witness both lie inside the difficulty's path range.
// No exponential board search or unverified fallback is necessary.
func Generate(ctx context.Context, seed uint64, d Difficulty) (Puzzle, error) {
	if d < 0 || d > 3 {
		return Puzzle{}, errors.New("invalid difficulty")
	}
	r := random(seed)
	rules := d.Rules()
	ranges := [][][2]int{{{0, 3}, {2, 5}}, {{0, 3}, {2, 6}, {4, 7}}, {{0, 4}, {3, 8}, {5, 9}}, {{0, 4}, {2, 6}, {4, 8}, {6, 10}}}[d]
	for attempt := 0; attempt < 512; attempt++ {
		if err := ctx.Err(); err != nil {
			return Puzzle{}, err
		}
		p := Puzzle{Version: Version, Seed: seed, Difficulty: d, Board: make([]string, rules.Size*rules.Size)}
		for i := range p.Board {
			p.Board[i] = Codes[r.next(len(Codes))]
		}
		s := New(p, Practice, true, "generate", time.Time{})
		length := ranges[len(ranges)-1][1]
		for len(s.Steps) < length {
			moves := s.Moves()
			if len(moves) == 0 {
				break
			}
			cell := moves[r.next(len(moves))]
			s.Steps = append(s.Steps, Step{Cell: cell})
			p.Solution = append(p.Solution, cell)
		}
		if len(p.Solution) != length {
			continue
		}
		sequence := s.Buffer()
		for _, span := range ranges {
			p.Targets = append(p.Targets, slices.Clone(sequence[span[0]:span[1]]))
		}
		contained := false
		for i, a := range p.Targets {
			for j, b := range p.Targets {
				if i != j {
					if ok, _ := Progress(a, b); ok {
						contained = true
					}
				}
			}
		}
		if contained || MinimumCodes(p.Targets) < rules.MinPath {
			continue
		}
		if err := p.Validate(); err != nil {
			continue
		}
		return p, nil
	}
	// A fixed, independently validated code stream is a deterministic fallback.
	// Its path uses a staircase with no repeated cells on every supported board.
	p := Puzzle{Version: Version, Seed: seed, Difficulty: d, Board: make([]string, rules.Size*rules.Size)}
	for i := range p.Board {
		p.Board[i] = Codes[r.next(len(Codes))]
	}
	seq := []string{"1C", "55", "BD", "E9", "7A", "FF", "1C", "BD", "55", "7A"}
	for i := 0; i < ranges[len(ranges)-1][1]; i++ {
		row := (i + 1) / 2
		col := i / 2
		cell := row*rules.Size + col
		p.Board[cell] = seq[i]
		p.Solution = append(p.Solution, cell)
	}
	for _, span := range ranges {
		p.Targets = append(p.Targets, slices.Clone(seq[span[0]:span[1]]))
	}
	if MinimumCodes(p.Targets) < rules.MinPath {
		return Puzzle{}, errors.New("fallback lower bound failed")
	}
	return p, p.Validate()
}

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

func TutorialPuzzle() Puzzle {
	return Puzzle{Version: Version, Seed: 2077, Difficulty: 1, Board: []string{
		"55", "1C", "E9", "BD", "7A", "FF", "BD", "E9", "55", "1C", "FF", "FF",
		"7A", "55", "FF", "E9", "BD", "1C", "FF", "BD", "1C", "55", "E9", "7A",
		"E9", "FF", "7A", "1C", "E9", "BD", "1C", "7A", "BD", "FF", "55", "E9",
	}, Targets: [][]string{{"1C", "55", "BD"}, {"BD", "E9", "7A"}, {"7A", "55", "FF"}}, Solution: []int{1, 13, 16, 28, 26, 8, 11}}
}
