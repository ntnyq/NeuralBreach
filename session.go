package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/locales"
)

func (w *window) advance(now time.Time) {
	if now.Before(w.lastTick) {
		return
	}
	if w.session != nil && w.page == gamePage && len(w.session.Steps) > 0 {
		ms := now.Sub(w.lastTick).Milliseconds()
		w.session.Advance(ms, now)
		w.lastTick = w.lastTick.Add(time.Duration(ms) * time.Millisecond)
		return
	}
	w.lastTick = now
}

func (w *window) pause() {
	if w.page != gamePage || w.session == nil {
		return
	}
	w.advance(time.Now())
	if w.session.Reason != "" {
		w.finish()
		return
	}
	w.navigate(pausePage)
	w.persist(nil)
}

func (w *window) resume() {
	if w.session == nil {
		return
	}
	w.session.Resumes++
	if w.session.Mode == game.Tutorial {
		w.focusCell = w.session.Puzzle.Solution[len(w.session.Steps)]
	}
	w.lastTick = time.Now()
	w.navigate(gamePage)
	w.focusBoard = true
	w.persist(nil)
}

func (w *window) selectCell(cell int) {
	if w.page != gamePage || w.session == nil || w.confirmAction != nil {
		return
	}
	w.advance(time.Now())
	if w.session.Mode == game.Tutorial && cell != w.session.Puzzle.Solution[len(w.session.Steps)] {
		w.audio.Effect("invalid")
		return
	}
	before := w.session.Uploaded()
	if !w.session.Select(cell, time.Now()) {
		w.audio.Effect("invalid")
		if w.session.Reason != "" {
			w.finish()
		}
		return
	}
	w.focusCell = cell
	w.hintCell = -1
	w.task.Cancel()
	w.message = nil
	if w.session.Uploaded() > before {
		w.audio.Effect("upload")
		w.triggerGlitch()
	} else {
		w.audio.Effect("select")
	}
	if w.session.Reason != "" {
		w.finish()
	} else {
		w.persist(nil)
	}
}

func (w *window) finish() {
	if w.session == nil || w.session.Reason == "" {
		return
	}
	s := *w.session
	w.result = &s
	w.data.Finish(s)
	w.session = nil
	w.task.Cancel()
	w.navigate(resultPage)
	w.persist(nil)
	if s.Reason == game.Complete {
		w.audio.Effect("complete")
	} else {
		w.audio.Effect("fail")
	}
}

func (w *window) start() {
	if w.loadErr != nil || w.task.Running() {
		return
	}
	if w.session != nil {
		w.ask(locales.AbandonConfirm, func() { w.abandon(); w.generate() })
		return
	}
	w.generate()
}

func (w *window) abandon() {
	if w.session == nil {
		return
	}
	w.session.End(game.Abandoned, time.Now())
	w.data.Finish(*w.session)
	w.session = nil
	w.persist(nil)
}

func (w *window) tutorial() {
	if w.session != nil {
		w.ask(locales.AbandonConfirm, func() { w.abandon(); w.tutorial() })
		return
	}
	s := game.New(game.TutorialPuzzle(), game.Tutorial, true, "tutorial", time.Now())
	w.session = &s
	w.focusCell = 1
	w.focusBoard = true
	w.hintCell = -1
	w.lastTick = time.Now()
	w.navigate(gamePage)
	w.persist(nil)
}

func (w *window) restart() {
	if w.session == nil {
		return
	}
	p := w.session.Puzzle
	w.ask(locales.AbandonConfirm, func() { w.abandon(); w.practicePuzzle(p) })
}

func (w *window) practicePuzzle(p game.Puzzle) {
	id, err := randomSeed()
	if err != nil {
		w.errorDetail = err
		return
	}
	s := game.New(p, game.Practice, true, fmt.Sprintf("%x-%x", time.Now().UnixNano(), id), time.Now())
	w.session = &s
	w.mode = game.Practice
	w.difficulty = p.Difficulty
	w.focusCell = 0
	w.focusBoard = true
	w.lastTick = time.Now()
	w.navigate(gamePage)
	w.persist(nil)
}

func (w *window) hint() {
	if w.session == nil || w.session.Mode != game.Practice || w.task.Running() {
		return
	}
	snapshot := *w.session
	snapshot.Steps = slices.Clone(snapshot.Steps)
	var path []int
	w.run(func(ctx context.Context) error {
		var err error
		path, err = game.Solve(ctx, snapshot, 250000)
		return err
	}, func() {
		if len(path) > 0 && w.session != nil {
			w.session.Hints++
			w.hintCell = path[0]
			w.focusCell = path[0]
			w.focusBoard = true
			w.persist(nil)
		}
	})
}
