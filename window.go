package main

import (
	"context"
	"errors"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/records"
	"github.com/ntnyq/cyber-game/internal/sound"
	"github.com/ntnyq/cyber-game/internal/tasks"
	"github.com/ntnyq/cyber-game/locales"
)

type page int

const (
	menuPage page = iota
	setupPage
	gamePage
	pausePage
	resultPage
	recordsPage
	replayPage
	settingsPage
)

type window struct {
	win                                                *mygo.Window
	postUpdate                                         func(func())
	store                                              *records.Store
	data                                               records.Data
	loadErr, saveErr, errorDetail                      error
	savedRevision                                      uint64
	page, settingsReturn                               page
	session                                            *game.Session
	result                                             *game.Session
	difficulty                                         game.Difficulty
	mode                                               game.Mode
	seed                                               string
	unlimited                                          bool
	draft                                              records.Settings
	audio                                              *sound.Engine
	previewAudio                                       bool
	background, closed, allowClose, quitting           bool
	done                                               chan struct{}
	task                                               tasks.Task
	lastTick                                           time.Time
	message                                            func() string
	confirmText                                        func() string
	confirmAction                                      func()
	menuIndex, focusCell, hintCell                     int
	focusBoard, focusMenu                              bool
	glitchStart, glitchLast                            time.Time
	replaySteps                                        int
	replaySolution, replayPlaying                      bool
	replayLast                                         time.Time
	recordIndex, filterMode, filterLevel, filterResult int
	listScroll, settingsScroll                         ui.ScrollState
}

func newWindow(store *records.Store, data records.Data, err error) *window {
	w := &window{store: store, data: data, loadErr: err, savedRevision: data.Revision, mode: game.Standard, difficulty: 1, hintCell: -1, done: make(chan struct{}), focusMenu: true, lastTick: time.Now()}
	if data.Active != nil {
		s := *data.Active
		w.session = &s
		w.menuIndex = 1
	}
	w.triggerGlitch()
	return w
}

func (w *window) dispatch(fn func()) {
	apply := func() {
		if !w.closed {
			fn()
		}
	}
	if w.postUpdate != nil {
		w.postUpdate(apply)
	} else {
		apply()
	}
}

func (w *window) loop() {
	ticker := time.NewTicker(33 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			w.dispatch(func() { w.tick(time.Now()) })
		}
	}
}

func (w *window) close() {
	if w.closed {
		return
	}
	w.closed = true
	close(w.done)
	w.task.Close()
	w.audio.Close()
}

func (w *window) tick(now time.Time) {
	if w.page == gamePage && w.confirmAction == nil && w.session != nil {
		w.advance(now)
		if w.session.Reason != "" {
			w.finish()
		}
	}
	if w.page == replayPage && w.replayPlaying && now.Sub(w.replayLast) >= 600*time.Millisecond {
		w.replayLast = now
		if w.replaySteps < len(w.replayPath()) {
			w.replaySteps++
		} else {
			w.replayPlaying = false
		}
	}
	w.updateAudio()
}

func (w *window) navigate(p page) {
	if p != w.page {
		w.task.Cancel()
	}
	w.page = p
	w.message = nil
	w.errorDetail = nil
	w.triggerGlitch()
	if p == menuPage {
		w.focusMenu = true
		w.menuIndex = 0
		if w.session != nil {
			w.menuIndex = 1
		}
	}
	w.updateAudio()
}

func (w *window) ask(text func() string, action func()) {
	if w.page == gamePage {
		w.pause()
		if w.page == resultPage {
			return
		}
	}
	w.confirmText = text
	w.confirmAction = action
}

func (w *window) cancelConfirm() { w.confirmAction = nil; w.confirmText = nil }

// run keeps blocking dialogs, solver work and exports off the main thread.
func (w *window) run(work func(context.Context) error, done func()) {
	w.task.Start(w.dispatch, func(ctx context.Context, _ func(float64)) (string, error) { return "", work(ctx) }, func(float64) {}, func(_ string, err error) {
		w.message = nil
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if errors.Is(err, game.ErrNoSolution) {
				w.message = locales.NoRoute
			} else if errors.Is(err, game.ErrBudget) {
				w.message = locales.HintUnavailable
			} else {
				w.errorDetail = err
			}
			return
		}
		done()
	})
}

func (w *window) requestQuit() {
	if w.quitting || w.confirmAction != nil {
		return
	}
	if w.session != nil || w.savedRevision != w.data.Revision {
		w.ask(locales.QuitConfirm, w.saveQuit)
	} else {
		w.allowClose = true
		if w.win != nil {
			mygo.App.Quit()
		}
	}
}

func (w *window) quitWithoutSave() {
	w.allowClose = true
	if w.win != nil {
		mygo.App.Quit()
	}
}
