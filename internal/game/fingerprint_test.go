package game

import (
	"slices"
	"testing"
)

func TestPuzzleFingerprint(t *testing.T) {
	p := TutorialPuzzle()
	want := p.Fingerprint()
	p.Seed++
	p.Solution = []int{0, 1}
	if p.Fingerprint() != want {
		t.Fatal("seed or witness changed content identity")
	}
	for name, change := range map[string]func(*Puzzle){
		"version":      func(p *Puzzle) { p.Version++ },
		"difficulty":   func(p *Puzzle) { p.Difficulty++ },
		"board":        func(p *Puzzle) { p.Board[0] = "FF" },
		"target":       func(p *Puzzle) { p.Targets[0][0] = "FF" },
		"target order": func(p *Puzzle) { slices.Reverse(p.Targets) },
		"target boundaries": func(p *Puzzle) {
			p.Targets[0] = append(p.Targets[0], p.Targets[1][0])
			p.Targets[1] = p.Targets[1][1:]
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := TutorialPuzzle()
			change(&p)
			if p.Fingerprint() == want {
				t.Fatal("different content has the same fingerprint")
			}
		})
	}
}
