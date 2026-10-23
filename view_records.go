package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/locales"
)

func (w *window) filteredRecords() []game.Session {
	var out []game.Session
	for _, s := range w.data.History {
		if w.filterMode == 1 && s.Mode != game.Standard || w.filterMode == 2 && s.Mode != game.Practice {
			continue
		}
		if w.filterLevel > 0 && int(s.Puzzle.Difficulty) != w.filterLevel-1 {
			continue
		}
		if w.filterResult == 1 && s.Reason != game.Complete || w.filterResult == 2 && (s.Uploaded() == 0 || s.Reason == game.Complete || s.Reason == game.Abandoned) || w.filterResult == 3 && (s.Uploaded() > 0 || s.Reason == game.Abandoned) || w.filterResult == 4 && s.Reason != game.Abandoned {
			continue
		}
		out = append(out, s)
	}
	return out
}

func (w *window) recordsView(c *ui.Context) {
	title(c, locales.Records())
	played, wins, high := w.data.Stats()
	rate := 0
	if played > 0 {
		rate = wins * 100 / played
	}
	ui.Text(c, fmt.Sprintf("%s %d     /     %s %d%%     /     %s %d", locales.CompletedGames(), played, locales.WinRate(), rate, locales.HighScore(), high)).TextColor(cyan).PaddingY(14)
	ui.Row(c).Gap(10).Children(func() {
		cycle(c, []string{locales.All(), locales.Standard(), locales.Practice()}, &w.filterMode)
		cycle(c, []string{locales.All(), levelName(0), levelName(1), levelName(2), levelName(3)}, &w.filterLevel)
		cycle(c, []string{locales.All(), locales.Success(), locales.Partial(), locales.Failure(), locales.Abandoned()}, &w.filterResult)
		ui.Spacer(c)
		button(c, locales.ExportJSON(), func() { w.export(false) })
		button(c, locales.ExportCSV(), func() { w.export(true) })
	})
	ui.Row(c).Grow(1).MinHeight(0).Gap(30).PaddingY(18).Children(func() {
		records := w.filteredRecords()
		ui.Scroll(c).WidthPercent(56).MinHeight(0).TrackScroll(&w.listScroll).Children(func() {
			if len(records) == 0 {
				ui.Text(c, locales.EmptyRecords()).Wrap()
				button(c, locales.StartGame(), func() { w.navigate(setupPage) })
			}
			for i, s := range records {
				label := fmt.Sprintf("%s    %s    %d\n%s / %s", s.FinishedAt.Local().Format("01-02 15:04"), levelName(s.Puzzle.Difficulty), s.Score().Total, modeName(s.Mode), resultName(s))
				b := ui.Button(c, label).Key(s.ID).Height(78).Justify(ui.Start).PaddingX(16)
				if i == w.recordIndex {
					b.Border(1, cyan).TextColor(cyan)
				}
				if b.Clicked() {
					w.recordIndex = i
				}
			}
		})
		ui.Column(c).Grow(1).Gap(15).Children(func() {
			sectionLabel(c, locales.PersonalBest())
			for d := game.Difficulty(0); d < 4; d++ {
				value := "—"
				if best := w.data.Best(d); best != nil {
					value = strconv.Itoa(best.Score().Total)
				}
				ui.Text(c, levelName(d)+"   /   "+value).TextColor(yellow)
			}
			if len(records) > 0 {
				w.recordIndex = min(w.recordIndex, len(records)-1)
				s := records[w.recordIndex]
				ui.Text(c, "SEED / "+strconv.FormatUint(s.Puzzle.Seed, 10)).TextColor(muted)
				ui.Text(c, fmt.Sprintf("%s %.1fs / %s %d", locales.Elapsed(), float64(s.ElapsedMS)/1000, locales.Resumes(), s.Resumes))
				button(c, locales.Replay(), func() { w.openReplay(s) })
				button(c, locales.Restart(), func() {
					if w.session != nil {
						w.ask(locales.AbandonConfirm, func() { w.abandon(); w.practicePuzzle(s.Puzzle) })
					} else {
						w.practicePuzzle(s.Puzzle)
					}
				})
			}
		})
	})
	button(c, locales.Back(), func() { w.navigate(menuPage) })
}

func (w *window) replayView(c *ui.Context) {
	if w.result == nil {
		return
	}
	ui.Row(c).Gap(10).Children(func() {
		button(c, locales.Back(), func() { w.replayPlaying = false; w.navigate(recordsPage) })
		button(c, locales.Previous(), func() { w.replaySteps = max(0, w.replaySteps-1) })
		button(c, locales.Next(), func() { w.replaySteps = min(len(w.replayPath()), w.replaySteps+1) })
		label := locales.Play()
		if w.replayPlaying {
			label = locales.Stop()
		}
		button(c, label, func() { w.replayPlaying = !w.replayPlaying; w.replayLast = time.Now() })
		ui.Text(c, fmt.Sprintf("%d / %d", w.replaySteps, len(w.replayPath())))
		ui.Spacer(c)
		label = locales.ActualPath()
		if w.replaySolution {
			label = locales.SolutionPath()
		}
		button(c, label, func() { w.replaySolution = !w.replaySolution; w.replaySteps = 0; w.replayPlaying = false })
	})
	w.breachView(c, w.replayState(), true)
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
