package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/sound"
	"github.com/ntnyq/cyber-game/locales"
)

func levelName(d game.Difficulty) string {
	return []func() string{locales.Easy, locales.Normal, locales.Hard, locales.Extreme}[d]()
}

func modeName(m game.Mode) string {
	switch m {
	case game.Practice:
		return locales.Practice()
	case game.Tutorial:
		return locales.Tutorial()
	default:
		return locales.Standard()
	}
}

func resultName(s game.Session) string {
	if s.Reason == game.Abandoned {
		return locales.Abandoned()
	}
	if s.Uploaded() == len(s.Puzzle.Targets) {
		return locales.Success()
	}
	if s.Uploaded() > 0 {
		return locales.Partial()
	}
	return locales.Failure()
}

func reasonName(r game.Reason) string {
	switch r {
	case game.Complete:
		return locales.Success()
	case game.Full:
		return locales.BufferFull()
	case game.Timeout:
		return locales.TimedOut()
	case game.Trapped:
		return locales.NoMoves()
	case game.Abandoned:
		return locales.Abandoned()
	default:
		return locales.ManualEnd()
	}
}

func (w *window) view(c *ui.Context) {
	scale := w.data.Settings.Scale
	c.SetTheme(gameTheme(scale))
	if w.quitting {
		ui.Column(c).Grow(1).Padding(40).Gap(20).Children(func() { title(c, locales.Saving()); w.statusView(c) })
		return
	}
	if c.Shortcut(0, ui.KeyM) {
		if w.page == settingsPage {
			w.draft.Muted = !w.draft.Muted
		} else {
			w.data.Settings.Muted = !w.data.Settings.Muted
			w.persist(nil)
		}
		w.updateAudio()
	}
	if w.confirmAction == nil && c.Shortcut(0, ui.KeyEscape) {
		switch w.page {
		case gamePage:
			w.pause()
		case pausePage:
			w.resume()
		case settingsPage:
			w.leaveSettings(false)
		case setupPage:
			w.task.Cancel()
			w.navigate(menuPage)
		case replayPage:
			w.replayPlaying = false
			w.navigate(recordsPage)
		case recordsPage, resultPage:
			w.navigate(menuPage)
		}
	}
	ui.Column(c.Key("shell")).Grow(1).Padding(20, 34).Gap(10).Background(black).Draw(w.drawBackground).DrawOver(w.drawGlitch).Children(func() {
		ui.Row(c).Shrink(0).AlignItems(ui.Center).Gap(16).Children(func() {
			ui.Text(c, "NEURAL / BREACH").FontSize(21).FontWeight(600).TextColor(red)
			ui.Spacer(c)
			ui.Text(c, locales.Offline()).TextColor(muted)
			ui.Text(c, "v"+version).TextColor(red)
		})
		ui.Box(c).Height(1).Shrink(0).Background(red.Alpha(0.55))
		ui.Column(c.Key(fmt.Sprintf("page-%d", w.page))).Grow(1).MinHeight(0).Children(func() {
			switch w.page {
			case menuPage:
				w.menuView(c)
			case setupPage:
				w.setupView(c)
			case gamePage:
				if w.session != nil {
					w.breachView(c, *w.session, false)
				}
			case pausePage:
				w.pauseView(c)
			case resultPage:
				w.resultView(c)
			case recordsPage:
				w.recordsView(c)
			case replayPage:
				w.replayView(c)
			case settingsPage:
				w.settingsView(c)
			}
		})
		w.statusView(c)
		ui.Row(c).Shrink(0).Gap(12).Children(func() {
			ui.Text(c, "◆  "+locales.Controls()).FontSize(13 * float32(scale) / 100).TextColor(muted)
			ui.Spacer(c)
			if w.data.Revision > 0 {
				label := locales.Saved()
				if w.savedRevision < w.data.Revision {
					label = locales.Saving()
				}
				ui.Text(c, label).FontSize(13).TextColor(cyan)
			}
			if w.data.Settings.Muted {
				ui.Text(c, "M / OFF").TextColor(red)
			}
		})
	})
	if w.confirmAction != nil {
		open := true
		var chosen bool
		ui.Modal(c.Key("confirmation"), &open, func() {
			title(c, locales.Confirm())
			ui.Text(c, w.confirmText()).Wrap()
			ui.Row(c).Gap(12).Children(func() {
				button(c, locales.Cancel(), func() { open = false })
				primary(c, locales.Confirm(), func() { chosen = true })
			})
		}).Width(560)
		if chosen {
			action := w.confirmAction
			w.cancelConfirm()
			action()
		} else if !open {
			w.cancelConfirm()
		}
	}
}

func (w *window) statusView(c *ui.Context) {
	if w.loadErr != nil {
		ui.Row(c).Gap(12).Children(func() {
			ui.Text(c, locales.DataDamaged()).TextColor(red)
			button(c, locales.Repair(), w.repair)
			button(c, locales.DiscardQuit(), func() { w.ask(locales.DiscardConfirm, w.quitWithoutSave) })
		})
		ui.Text(c, w.loadErr.Error()).TextColor(muted).FontSize(12)
	} else if w.saveErr != nil {
		ui.Row(c).Gap(10).Children(func() {
			ui.Text(c, locales.SaveFailed()).TextColor(red)
			button(c, locales.RetrySave(), func() { w.persist(nil) })
			button(c, locales.DiscardQuit(), func() { w.ask(locales.DiscardConfirm, w.quitWithoutSave) })
		})
		ui.Text(c, w.saveErr.Error()).TextColor(muted).FontSize(12)
	}
	if w.errorDetail != nil {
		ui.Text(c, w.errorDetail.Error()).TextColor(red).Wrap()
	}
	if w.message != nil {
		ui.Text(c, w.message()).TextColor(cyan)
	}
	if w.task.Running() {
		ui.Text(c, locales.Busy()).TextColor(muted)
	}
	if sound.Error() != nil {
		ui.Text(c, locales.AudioUnavailable()).TextColor(muted)
	}
}
