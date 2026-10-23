package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/records"
	"github.com/ntnyq/cyber-game/internal/sound"
	"github.com/ntnyq/cyber-game/internal/storage"
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
func (w *window) triggerGlitch() {
	now := time.Now()
	if now.Sub(w.glitchLast) < 1500*time.Millisecond {
		return
	}
	w.glitchStart = now
	w.glitchLast = now
	if w.data.Settings.Glitch > 0 && !w.data.Settings.Reduced {
		w.audio.Effect("glitch")
	}
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
func (w *window) persist(after func()) {
	if w.store == nil {
		if after != nil {
			after()
		}
		return
	}
	if w.session != nil {
		s := *w.session
		w.data.Active = &s
		if s.Mode != game.Tutorial {
			w.data.RememberPuzzle(s.Puzzle)
		}
	}
	w.data.Revision++
	snapshot := w.data.Clone()
	rev := snapshot.Revision
	w.saveErr = nil
	go func() {
		err := w.store.Commit(snapshot)
		w.dispatch(func() {
			if rev != w.data.Revision {
				return
			}
			w.saveErr = err
			if err == nil {
				w.savedRevision = rev
				if after != nil {
					after()
				}
			} else {
				w.quitting = false
			}
		})
	}()
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
func randomSeed() (uint64, error) {
	var b [8]byte
	_, err := rand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:]), err
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
func (w *window) generate() {
	var seed *uint64
	if w.mode == game.Practice && w.seed != "" {
		value, err := strconv.ParseUint(w.seed, 10, 64)
		if err != nil {
			w.message = locales.InvalidSeed
			return
		}
		seed = &value
	}
	id, err := randomSeed()
	if err != nil {
		w.errorDetail = err
		return
	}
	level, mode, unlimited := w.difficulty, w.mode, w.unlimited
	// Include unsaved sessions without reading mutable UI state in the worker.
	seen := maps.Clone(w.data.PuzzleFingerprints)
	store := w.store
	reserve := func(ctx context.Context, p game.Puzzle) (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if seen[p.Fingerprint()] {
			return false, nil
		}
		if store != nil {
			return store.ReservePuzzle(ctx, p)
		}
		return true, nil
	}
	var puzzle game.Puzzle
	w.message = locales.Generating
	w.run(func(ctx context.Context) error {
		var err error
		puzzle, err = generatePuzzle(ctx, level, seed, randomSeed, reserve)
		return err
	}, func() {
		s := game.New(puzzle, mode, unlimited, fmt.Sprintf("%x-%x", time.Now().UnixNano(), id), time.Now())
		w.data.RememberPuzzle(puzzle)
		w.session = &s
		w.focusCell = 0
		w.focusBoard = true
		w.hintCell = -1
		w.lastTick = time.Now()
		w.navigate(gamePage)
		w.persist(nil)
	})
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
func (w *window) openSettings() {
	if w.page == gamePage {
		w.pause()
	}
	w.settingsReturn = w.page
	w.draft = w.data.Settings
	w.previewAudio = false
	w.navigate(settingsPage)
}
func (w *window) leaveSettings(apply bool) {
	if apply {
		w.data.Settings = w.draft
		if w.win != nil {
			mygo.App.SetLocale(w.data.Settings.Language)
			w.applyScale()
			w.syncMenu()
		}
		w.persist(nil)
	}
	w.previewAudio = false
	w.navigate(w.settingsReturn)
}
func (w *window) applyScale() {
	if w.win != nil {
		scale := float64(w.data.Settings.Scale) / 100
		w.win.SetMinimumSize(int(1100*scale), int(700*scale))
	}
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
func (w *window) saveQuit() {
	w.quitting = true
	w.persist(func() {
		w.allowClose = true
		if w.win != nil {
			mygo.App.Quit()
		}
	})
}
func (w *window) quitWithoutSave() {
	w.allowClose = true
	if w.win != nil {
		mygo.App.Quit()
	}
}
func (w *window) repair() {
	w.ask(locales.RepairConfirm, func() {
		var d records.Data
		w.run(func(context.Context) error { var err error; d, err = w.store.Repair(); return err }, func() {
			w.data = d
			w.loadErr = nil
			w.saveErr = nil
			w.session = nil
			w.savedRevision = d.Revision
			w.navigate(menuPage)
		})
	})
}
func (w *window) export(csv bool) {
	snapshot := w.data.Clone()
	w.run(func(ctx context.Context) error {
		ext := "json"
		if csv {
			ext = "csv"
		}
		path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{Parent: w.win, Title: appName, DefaultPath: "neural-breach." + ext})
		if err != nil {
			return err
		}
		if path == "" {
			return context.Canceled
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		var b []byte
		if csv {
			b, err = records.CSV(snapshot.History)
		} else {
			b, err = json.MarshalIndent(snapshot.History, "", "  ")
		}
		if err != nil {
			return err
		}
		return storage.WriteFile(path, b, 0o600)
	}, func() { w.message = locales.Exported })
}
func (w *window) openReplay(s game.Session) {
	w.result = &s
	w.replaySteps = 0
	w.replaySolution = false
	w.replayPlaying = false
	w.navigate(replayPage)
}
func (w *window) replayPath() []int {
	if w.result == nil {
		return nil
	}
	if w.replaySolution {
		return w.result.Puzzle.Solution
	}
	p := make([]int, len(w.result.Steps))
	for i, s := range w.result.Steps {
		p[i] = s.Cell
	}
	return p
}
func (w *window) replayState() game.Session {
	s := game.New(w.result.Puzzle, game.Practice, true, "replay", w.result.StartedAt)
	for _, cell := range w.replayPath()[:min(w.replaySteps, len(w.replayPath()))] {
		s.Select(cell, time.Now())
	}
	return s
}
func (w *window) updateAudio() {
	settings := w.data.Settings
	if w.page == settingsPage {
		settings = w.draft
	}
	stage := sound.Menu
	switch w.page {
	case gamePage:
		stage = sound.Observe
		if w.session != nil && len(w.session.Steps) > 0 {
			stage = sound.Breach
			if !w.session.Unlimited && w.session.Remaining() <= 10000 {
				stage = sound.Urgent
			}
		}
	case pausePage:
		stage = sound.Paused
	case settingsPage:
		if w.settingsReturn == pausePage && !w.previewAudio {
			stage = sound.Paused
		}
		if w.previewAudio {
			stage = sound.Breach
		}
	case resultPage:
		stage = sound.Result
	}
	if w.background && settings.BackgroundMute {
		stage = sound.Paused
	}
	w.audio.Update(stage, settings)
}
