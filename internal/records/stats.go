package records

import (
	"github.com/ntnyq/cyber-game/internal/game"
)

func Better(a, b game.Session) bool {
	if a.Score().Total != b.Score().Total {
		return a.Score().Total > b.Score().Total
	}
	if a.Uploaded() != b.Uploaded() {
		return a.Uploaded() > b.Uploaded()
	}
	if len(a.Steps) != len(b.Steps) {
		return len(a.Steps) < len(b.Steps)
	}
	if a.ElapsedMS != b.ElapsedMS {
		return a.ElapsedMS < b.ElapsedMS
	}
	return a.FinishedAt.Before(b.FinishedAt)
}

func (d Data) Best(level game.Difficulty) *game.Session {
	var best *game.Session
	for _, s := range d.History {
		if s.Mode != game.Standard || s.Reason == game.Abandoned || s.Puzzle.Difficulty != level || s.Puzzle.Version != game.Version {
			continue
		}
		if best == nil || Better(s, *best) {
			copy := s
			best = &copy
		}
	}
	return best
}

func (d Data) Stats() (played, wins, high int) {
	for _, s := range d.History {
		if s.Mode != game.Standard || s.Reason == game.Abandoned {
			continue
		}
		played++
		if s.Reason == game.Complete {
			wins++
		}
		high = max(high, s.Score().Total)
	}
	return
}
