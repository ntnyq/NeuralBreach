// Package game implements the deterministic, UI-independent breach rules.
package game

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

const Version = 1

var Codes = []string{"1C", "55", "BD", "E9", "7A", "FF"}

type Difficulty int
type Rules struct{ Size, Capacity, Targets, Seconds, Multiplier, MinPath, MaxPath int }

func (d Difficulty) Rules() Rules {
	return [...]Rules{{5, 7, 2, 90, 2, 5, 6}, {6, 8, 3, 60, 3, 6, 7}, {7, 9, 3, 45, 4, 8, 9}, {8, 10, 4, 35, 6, 10, 10}}[d]
}

type Mode string

const (
	Standard Mode = "standard"
	Practice Mode = "practice"
	Tutorial Mode = "tutorial"
)

type Reason string

const (
	Complete  Reason = "complete"
	Full      Reason = "buffer"
	Timeout   Reason = "timeout"
	Trapped   Reason = "trapped"
	Manual    Reason = "manual"
	Abandoned Reason = "abandoned"
)

type Puzzle struct {
	Version    int        `json:"version"`
	Seed       uint64     `json:"seed"`
	Difficulty Difficulty `json:"difficulty"`
	Board      []string   `json:"board"`
	Targets    [][]string `json:"targets"`
	Solution   []int      `json:"solution"`
}
type Step struct {
	Cell      int   `json:"cell"`
	ElapsedMS int64 `json:"elapsedMs"`
}
type Session struct {
	ID         string    `json:"id"`
	Puzzle     Puzzle    `json:"puzzle"`
	Mode       Mode      `json:"mode"`
	Unlimited  bool      `json:"unlimited"`
	Steps      []Step    `json:"steps"`
	ElapsedMS  int64     `json:"elapsedMs"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt,omitempty"`
	Reason     Reason    `json:"reason,omitempty"`
	Resumes    int       `json:"resumes"`
	Hints      int       `json:"hints"`
	Undos      int       `json:"undos"`
}

func New(p Puzzle, mode Mode, unlimited bool, id string, now time.Time) Session {
	return Session{ID: id, Puzzle: p, Mode: mode, Unlimited: unlimited && mode != Standard, StartedAt: now}
}
func (s Session) Buffer() []string {
	b := make([]string, len(s.Steps))
	for i, v := range s.Steps {
		b[i] = s.Puzzle.Board[v.Cell]
	}
	return b
}
func Progress(buffer, target []string) (bool, int) {
	for i := 0; i+len(target) <= len(buffer); i++ {
		if slices.Equal(buffer[i:i+len(target)], target) {
			return true, len(target)
		}
	}
	for n := min(len(buffer), len(target)-1); n > 0; n-- {
		if slices.Equal(buffer[len(buffer)-n:], target[:n]) {
			return false, n
		}
	}
	return false, 0
}
func (s Session) Uploaded() int {
	n := 0
	for _, target := range s.Puzzle.Targets {
		ok, _ := Progress(s.Buffer(), target)
		if ok {
			n++
		}
	}
	return n
}
func (s Session) Used(cell int) bool {
	for _, step := range s.Steps {
		if step.Cell == cell {
			return true
		}
	}
	return false
}
func (s Session) Legal(cell int) bool {
	n := s.Puzzle.Difficulty.Rules().Size
	if cell < 0 || cell >= len(s.Puzzle.Board) || s.Used(cell) {
		return false
	}
	if len(s.Steps) == 0 {
		return cell < n
	}
	last := s.Steps[len(s.Steps)-1].Cell
	if len(s.Steps)%2 == 1 {
		return cell%n == last%n
	}
	return cell/n == last/n
}
func (s Session) Moves() []int {
	var out []int
	for cell := range s.Puzzle.Board {
		if s.Legal(cell) {
			out = append(out, cell)
		}
	}
	return out
}
func (s Session) Remaining() int64 {
	return max(0, int64(s.Puzzle.Difficulty.Rules().Seconds)*1000-s.ElapsedMS)
}
func (s *Session) End(reason Reason, now time.Time) {
	if s.Reason == "" {
		s.Reason = reason
		s.FinishedAt = now
	}
}
func (s *Session) Advance(ms int64, now time.Time) {
	if s.Reason != "" || len(s.Steps) == 0 {
		return
	}
	s.ElapsedMS += max(ms, 0)
	if !s.Unlimited && s.Remaining() == 0 {
		s.ElapsedMS = int64(s.Puzzle.Difficulty.Rules().Seconds) * 1000
		s.End(Timeout, now)
	}
}
func (s *Session) Select(cell int, now time.Time) bool {
	if s.Reason != "" {
		return false
	}
	if !s.Unlimited && s.Remaining() == 0 {
		s.End(Timeout, now)
		return false
	}
	if !s.Legal(cell) {
		return false
	}
	s.Steps = append(s.Steps, Step{cell, s.ElapsedMS})
	switch {
	case s.Uploaded() == len(s.Puzzle.Targets):
		s.End(Complete, now)
	case len(s.Steps) == s.Puzzle.Difficulty.Rules().Capacity:
		s.End(Full, now)
	case len(s.Moves()) == 0:
		s.End(Trapped, now)
	}
	return true
}
func (s *Session) Undo() bool {
	if s.Mode != Practice || len(s.Steps) == 0 || s.Reason != "" {
		return false
	}
	s.Steps = s.Steps[:len(s.Steps)-1]
	s.Undos++
	return true
}

type Score struct{ Base, Combo, Complete, Time, Buffer, Total int }

func (s Session) Score() Score {
	var score Score
	k := 0
	for _, target := range s.Puzzle.Targets {
		if ok, _ := Progress(s.Buffer(), target); ok {
			score.Base += len(target) * 100
			k++
		}
	}
	if k == 0 || s.Reason == Abandoned {
		return Score{}
	}
	score.Combo = 150 * max(k-1, 0)
	if k == len(s.Puzzle.Targets) {
		score.Complete = 400
		score.Buffer = 80 * (s.Puzzle.Difficulty.Rules().Capacity - len(s.Steps))
		if !s.Unlimited {
			score.Time = int(300 * s.Remaining() / int64(s.Puzzle.Difficulty.Rules().Seconds*1000))
		}
	}
	score.Total = (score.Base + score.Combo + score.Complete + score.Time + score.Buffer) * s.Puzzle.Difficulty.Rules().Multiplier / 2
	return score
}
func (s Session) Validate() error {
	if err := s.Puzzle.Validate(); err != nil {
		return err
	}
	if s.ID == "" || s.StartedAt.IsZero() || s.ElapsedMS < 0 || s.Resumes < 0 || s.Hints < 0 || s.Undos < 0 {
		return errors.New("invalid session metadata")
	}
	if s.Mode != Standard && s.Mode != Practice && s.Mode != Tutorial {
		return errors.New("unknown mode")
	}
	if s.Mode == Standard && (s.Unlimited || s.Hints > 0 || s.Undos > 0) {
		return errors.New("assisted standard game")
	}
	if s.Mode == Tutorial && (!s.Unlimited || s.Hints > 0 || s.Undos > 0) {
		return errors.New("invalid tutorial rules")
	}
	if !s.Unlimited && s.ElapsedMS > int64(s.Puzzle.Difficulty.Rules().Seconds*1000) {
		return errors.New("invalid elapsed time")
	}
	copy := New(s.Puzzle, s.Mode, s.Unlimited, s.ID, s.StartedAt)
	var last int64
	for i, step := range s.Steps {
		if s.Mode == Tutorial && (i >= len(s.Puzzle.Solution) || step.Cell != s.Puzzle.Solution[i]) {
			return errors.New("invalid tutorial path")
		}
		if step.ElapsedMS < last || step.ElapsedMS > s.ElapsedMS {
			return errors.New("invalid step time")
		}
		copy.ElapsedMS = step.ElapsedMS
		if !copy.Select(step.Cell, s.StartedAt) {
			return errors.New("invalid path")
		}
		last = step.ElapsedMS
	}
	if copy.Reason != "" && copy.Reason != s.Reason {
		return errors.New("incorrect ending")
	}
	if s.Reason == Complete && copy.Reason != Complete {
		return errors.New("incomplete upload")
	}
	switch s.Reason {
	case "", Complete, Full, Trapped, Manual, Abandoned:
	case Timeout:
		if s.Unlimited || s.Remaining() != 0 {
			return errors.New("invalid timeout")
		}
	default:
		return errors.New("unknown ending")
	}
	if (s.Reason == Full || s.Reason == Trapped) && copy.Reason != s.Reason {
		return errors.New("invalid ending")
	}
	if s.Reason == "" && !s.Unlimited && s.Remaining() == 0 {
		return errors.New("expired active game")
	}
	if (s.Reason == "") != s.FinishedAt.IsZero() {
		return errors.New("invalid finish timestamp")
	}
	return nil
}
func (p Puzzle) Validate() error {
	if p.Version != Version || p.Difficulty < 0 || p.Difficulty > 3 {
		return errors.New("unsupported puzzle version or difficulty")
	}
	r := p.Difficulty.Rules()
	if len(p.Board) != r.Size*r.Size || len(p.Targets) != r.Targets {
		return errors.New("invalid puzzle dimensions")
	}
	for _, code := range p.Board {
		if !slices.Contains(Codes, code) {
			return fmt.Errorf("unknown code %q", code)
		}
	}
	for _, target := range p.Targets {
		if len(target) < 3 || len(target) > 6 {
			return errors.New("invalid target length")
		}
		for _, code := range target {
			if !slices.Contains(Codes, code) {
				return errors.New("invalid target code")
			}
		}
	}
	if len(p.Solution) > r.Capacity || len(p.Solution) == 0 {
		return errors.New("missing solution")
	}
	s := New(p, Practice, true, "verify", time.Now())
	for _, cell := range p.Solution {
		if !s.Select(cell, time.Now()) {
			return errors.New("invalid solution path")
		}
	}
	if s.Reason != Complete {
		return errors.New("solution does not upload all targets")
	}
	return nil
}
