package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/internal/sound"
	"github.com/ntnyq/cyber-game/locales"
)

var (
	black  = ui.Hex("#100C10")
	panel  = ui.Hex("#211219")
	red    = ui.Hex("#F25B64")
	yellow = ui.Hex("#FCEE09")
	cyan   = ui.Hex("#75F0E7")
	ivory  = ui.Hex("#E6DDE1")
	muted  = ui.Hex("#AB929C")
)

func gameTheme(scale int) *ui.Theme {
	t := ui.DarkTheme()
	t.Background = black
	t.Surface = panel
	t.SurfaceHover = ui.Hex("#382029")
	t.SurfacePressed = ui.Hex("#4A2631")
	t.Text = ivory
	t.TextMuted = muted
	t.Border = red.Alpha(0.4)
	t.Accent = cyan
	t.AccentText = black
	t.Focus = cyan
	t.Danger = red
	t.Selection = cyan.Alpha(0.25)
	t.Radius = 0
	t.FontSize = 17 * float32(scale) / 100
	t.Font = "Rajdhani"
	return t
}
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
func button(c *ui.Context, label string, action func()) {
	if ui.Button(c, label).Padding(10, 18).Clicked() {
		action()
	}
}
func primary(c *ui.Context, label string, action func()) {
	if ui.Button(c, label).Padding(12, 24).Background(yellow).TextColor(black).FontWeight(600).Clicked() {
		action()
	}
}
func title(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(c.Theme().FontSize * 2).FontWeight(600).TextColor(yellow)
}
func sectionLabel(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(c.Theme().FontSize).TextColor(red)
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
func (w *window) menuView(c *ui.Context) {
	commands := w.commands()
	for _, move := range []struct {
		key   ui.Key
		delta int
	}{{ui.KeyDown, 1}, {ui.KeyUp, -1}} {
		if c.Shortcut(0, move.key) {
			for range len(commands) {
				w.menuIndex = (w.menuIndex + move.delta + len(commands)) % len(commands)
				if commands[w.menuIndex].enabled {
					break
				}
			}
			w.focusMenu = true
			w.audio.Effect("move")
		}
	}
	ui.Row(c).Grow(1).Gap(70).AlignItems(ui.Center).Children(func() {
		ui.Column(c).WidthPercent(45).Gap(20).Children(func() {
			ui.Text(c, locales.GameTitle()).FontSize(52).FontWeight(600).TextColor(yellow).Skew(-7, 0)
			ui.Text(c, "NEURAL BREACH").FontSize(30).FontWeight(600).TextColor(yellow).Skew(-7, 0)
			ui.Text(c, locales.Tagline()).TextColor(muted)
			ui.Column(c).Gap(3).PaddingY(8).Children(func() {
				for i, cmd := range commands {
					b := ui.Button(c, cmd.label()).Key(fmt.Sprintf("menu-%d", i)).Height(48).FontSize(25*float32(w.data.Settings.Scale)/100).FontWeight(400).Justify(ui.Start).PaddingX(18).Background(black.Alpha(0.6)).TextColor(red).Border(1, red.Alpha(0.2)).Disabled(!cmd.enabled)
					if i == w.menuIndex {
						b.Background(cyan.Alpha(0.14)).TextColor(cyan).Border(1, cyan)
						if w.focusMenu {
							b.Focus()
						}
					}
					if b.Focused() && !w.focusMenu {
						w.menuIndex = i
					}
					if b.Clicked() {
						w.menuIndex = i
						cmd.run()
					}
				}
			})
			w.focusMenu = false
			if w.session == nil {
				ui.Text(c, locales.NoSave()).TextColor(muted)
			} else {
				ui.Text(c, levelName(w.session.Puzzle.Difficulty)+" / "+modeName(w.session.Mode)+fmt.Sprintf(" / %d/%d", len(w.session.Steps), w.session.Puzzle.Difficulty.Rules().Capacity)).TextColor(cyan)
			}
		})
		ui.Spacer(c)
	})
}
func (w *window) drawBackground(p *ui.Painter, r ui.Rect) {
	if w.page != menuPage || menuBackground == nil {
		return
	}
	p.Image(menuBackground, r, ui.Cover)
	p.FillGradient(r, ui.LinearGradient{Angle: 90, Stops: []ui.GradientStop{{Color: black.Alpha(0.45), At: 0}, {Color: black.Alpha(0.2), At: 0.42}, {Color: black.Alpha(0.04), At: 1}}}, 0)
}
func (w *window) setupView(c *ui.Context) {
	title(c, locales.StartGame())
	ui.Row(c).Gap(12).PaddingY(18).Children(func() {
		for _, m := range []game.Mode{game.Standard, game.Practice} {
			b := ui.Button(c, modeName(m)).Padding(12, 20)
			if w.mode == m {
				b.Background(cyan).TextColor(black)
			}
			if b.Clicked() {
				w.mode = m
			}
		}
		button(c, locales.Tutorial(), w.tutorial)
	})
	ui.Row(c).Grow(1).Gap(40).Children(func() {
		ui.Column(c).WidthPercent(40).Gap(12).Children(func() {
			for d := game.Difficulty(0); d < 4; d++ {
				b := ui.Button(c, fmt.Sprintf("%02d   %s", d+1, levelName(d))).Height(74).FontSize(25).Justify(ui.Start).PaddingX(24).TextColor(red)
				if w.difficulty == d {
					b.Background(cyan).TextColor(black)
				}
				if b.Clicked() {
					w.difficulty = d
					w.audio.Effect("move")
				}
			}
		})
		ui.Column(c).Grow(1).Gap(22).Children(func() {
			r := w.difficulty.Rules()
			title(c, levelName(w.difficulty))
			ui.Text(c, fmt.Sprintf("%d × %d  /  %d s  /  ×%.1f", r.Size, r.Size, r.Seconds, float64(r.Multiplier)/2)).FontSize(42).TextColor(yellow)
			ui.Text(c, fmt.Sprintf("%s  %d     %s  %d", locales.Buffer(), r.Capacity, locales.Targets(), r.Targets)).FontSize(21)
			if w.mode == game.Practice {
				ui.Text(c, locales.PracticeInfo()).Wrap()
				ui.TextInput(c.Key("seed"), &w.seed).Label(locales.Seed()).Placeholder(locales.Seed())
				ui.Checkbox(c, &w.unlimited, locales.Unlimited())
			} else {
				ui.Text(c, locales.StandardInfo()).Wrap()
			}
			ui.Spacer(c)
			if !w.task.Running() {
				primary(c, locales.Connect(), w.start)
			} else {
				ui.Text(c, locales.Generating()).TextColor(cyan)
			}
		})
	})
	button(c, locales.Back(), func() { w.task.Cancel(); w.navigate(menuPage) })
}
func (w *window) breachView(c *ui.Context, s game.Session, replay bool) {
	r := s.Puzzle.Difficulty.Rules()
	scale := float32(w.data.Settings.Scale) / 100
	ui.Row(c).Height(44 * scale).Shrink(0).AlignItems(ui.Center).Gap(18).Children(func() {
		title(c, locales.Matrix())
		ui.Spacer(c)
		ui.Text(c, levelName(s.Puzzle.Difficulty)+" / "+modeName(s.Mode)).TextColor(red)
		if !replay {
			button(c, locales.Pause(), w.pause)
		}
	})
	ui.Row(c).Height(78 * scale).Shrink(0).Gap(30).Children(func() {
		ui.Column(c).WidthPercent(48).Gap(5).Children(func() {
			sectionLabel(c, locales.TimeLeft())
			remaining := fmt.Sprintf("%05.1f", float64(s.Remaining())/1000)
			if s.Unlimited {
				remaining = "∞"
			}
			color := yellow
			if !s.Unlimited && s.Remaining() <= 10000 {
				color = red
			}
			ui.Text(c, remaining).FontSize(38 * scale).TextColor(color)
		})
		ui.Column(c).Grow(1).Gap(8).Children(func() {
			sectionLabel(c, fmt.Sprintf("%s  %d/%d", locales.Buffer(), len(s.Steps), r.Capacity))
			ui.Row(c).Gap(4).Children(func() {
				b := s.Buffer()
				for i := 0; i < r.Capacity; i++ {
					code := "—"
					col := muted
					if i < len(b) {
						code = b[i]
						col = cyan
					}
					ui.Text(c, code).Width(39*scale).Height(40*scale).FontSize(23*scale).TextAlign(ui.Center).PaddingY(5).Border(1, col.Alpha(0.5)).TextColor(col)
				}
			})
		})
	})
	if s.Mode == game.Tutorial {
		message := locales.TrainingOverlap
		switch len(s.Steps) {
		case 0:
			message = locales.TrainingFirst
		case 1:
			message = locales.TrainingAlternate
		case 2:
			message = locales.TrainingMatch
		}
		ui.Text(c, message()).TextColor(cyan).PaddingY(8)
	}
	_, screenHeight := c.Size()
	bodyHeight := screenHeight - 292*scale
	if replay {
		bodyHeight -= 44 * scale
	}
	if s.Mode == game.Tutorial {
		bodyHeight -= 30 * scale
	}
	ui.Row(c).Height(bodyHeight).Shrink(0).AlignItems(ui.Start).Gap(32).Children(func() {
		ui.Column(c).WidthPercent(48).Gap(10).Children(func() {
			label := locales.Observe()
			if len(s.Steps) > 0 {
				label = locales.ChooseRow()
				if len(s.Steps)%2 == 1 {
					label = locales.ChooseColumn()
				}
			}
			ui.Text(c, label).TextColor(muted).FontSize(14 * scale)
			w.board(c, s, replay, scale)
		})
		ui.Scroll(c).Grow(1).Height(bodyHeight).MinHeight(0).Children(func() {
			ui.Column(c).Gap(16).Children(func() {
				sectionLabel(c, locales.Targets())
				names := []func() string{locales.Siphon, locales.Silence, locales.Erase, locales.Ghost}
				for i, target := range s.Puzzle.Targets {
					ok, k := game.Progress(s.Buffer(), target)
					col := red
					state := locales.Waiting()
					if k > 0 {
						col = yellow
						state = fmt.Sprintf("%s %d/%d", locales.Matching(), k, len(target))
					}
					if ok {
						col = cyan
						state = "✓ " + locales.Uploaded()
					}
					ui.Column(c).Shrink(0).Padding(11, 17).Gap(5).Border(1, col.Alpha(0.45)).Background(panel).Children(func() {
						ui.Text(c, names[i]()).TextColor(col).FontSize(19 * scale)
						ui.Text(c, strings.Join(target, "  ")).FontSize(27 * scale).TextColor(yellow)
						ui.Text(c, state).FontSize(14 * scale).TextColor(col)
					})
				}
			})
		})
	})
	if !replay {
		ui.Row(c).Shrink(0).Gap(10).PaddingY(4).Children(func() {
			if s.Mode == game.Practice {
				button(c, locales.Hint(), w.hint)
				if ui.Button(c, locales.Undo()).Disabled(len(s.Steps) == 0).Clicked() {
					w.advance(time.Now())
					if w.session.Reason != "" {
						w.finish()
					} else if w.session.Undo() {
						w.task.Cancel()
						w.hintCell = -1
						w.persist(nil)
					}
				}
			}
			ui.Spacer(c)
			button(c, locales.Finish(), func() { w.ask(locales.FinishConfirm, func() { w.session.End(game.Manual, time.Now()); w.finish() }) })
		})
	}
}
func (w *window) board(c *ui.Context, s game.Session, replay bool, scale float32) {
	n := s.Puzzle.Difficulty.Rules().Size
	_, height := c.Size()
	cell := float32(math.Floor(float64((height - 355*scale) / float32(n))))
	if s.Mode == game.Tutorial || replay {
		cell -= 8 * scale
	}
	cell = max(44*scale, min(64*scale, cell))
	cell = min(cell, 51*scale*8/float32(n))
	if !replay {
		for _, move := range []struct {
			key   ui.Key
			delta int
		}{{ui.KeyLeft, -1}, {ui.KeyRight, 1}, {ui.KeyUp, -n}, {ui.KeyDown, n}} {
			if c.Shortcut(0, move.key) {
				w.focusCell = (w.focusCell + move.delta + n*n) % (n * n)
				w.focusBoard = true
				w.audio.Effect("move")
			}
		}
	}
	ui.Column(c.Key("board")).Shrink(0).Gap(4).Children(func() {
		for row := 0; row < n; row++ {
			ui.Row(c).Shrink(0).Gap(4).Children(func() {
				for col := 0; col < n; col++ {
					index := row*n + col
					code := s.Puzzle.Board[index]
					used := s.Used(index)
					legal := s.Legal(index)
					color, bg := yellow.Alpha(0.48), black
					if legal {
						color = yellow
						bg = yellow.Alpha(0.09)
					}
					if used {
						color = muted
						bg = panel
					}
					if index == w.hintCell && !replay {
						color = cyan
						bg = cyan.Alpha(0.16)
					}
					if s.Mode == game.Tutorial && len(s.Steps) < len(s.Puzzle.Solution) && index == s.Puzzle.Solution[len(s.Steps)] {
						color = cyan
						bg = cyan.Alpha(0.17)
					}
					b := ui.Button(c, code).Key(fmt.Sprintf("cell-%d", index)).Label(fmt.Sprintf("R%d C%d %s", row+1, col+1, code)).Size(cell, cell).Padding(0).FontSize(26*scale).TextColor(color).Background(bg).Border(1, color.Alpha(0.35))
					if used {
						step := 0
						for i, v := range s.Steps {
							if v.Cell == index {
								step = i + 1
							}
						}
						b.DrawOver(func(p *ui.Painter, r ui.Rect) {
							p.RichText(r.X+4, r.Y+2, 20, ui.Span{Text: strconv.Itoa(step), Size: 10 * scale, Color: cyan, Font: "Rajdhani"})
						})
					}
					if !replay && w.focusBoard && w.focusCell == index {
						b.Focus()
					}
					if !replay && b.Focused() && !w.focusBoard {
						w.focusCell = index
					}
					if b.Clicked() && !replay {
						w.selectCell(index)
					}
				}
			})
		}
	})
	if !replay {
		w.focusBoard = false
	}
}
func (w *window) pauseView(c *ui.Context) {
	ui.Column(c).Grow(1).Justify(ui.Center).AlignItems(ui.Center).Gap(22).Children(func() {
		ui.Text(c, "// CONNECTION SUSPENDED").TextColor(red)
		title(c, locales.Paused())
		ui.Text(c, locales.PauseInfo()).TextColor(muted)
		primary(c, locales.Resume(), w.resume)
		button(c, locales.SaveMenu(), func() { w.persist(func() { w.navigate(menuPage) }) })
		button(c, locales.Settings(), w.openSettings)
		button(c, locales.Restart(), w.restart)
		button(c, locales.QuitGame(), w.requestQuit)
	})
}
func (w *window) resultView(c *ui.Context) {
	if w.result == nil {
		return
	}
	s := *w.result
	score := s.Score()
	ui.Row(c).Grow(1).Gap(50).Children(func() {
		ui.Column(c).WidthPercent(48).Gap(20).Justify(ui.Center).Children(func() {
			title(c, resultName(s))
			label := locales.Score()
			if s.Mode == game.Practice {
				label = locales.PracticeScore()
			}
			ui.Text(c, label).TextColor(red)
			ui.Text(c, strconv.Itoa(score.Total)).FontSize(100).FontWeight(600).TextColor(cyan)
			ui.Text(c, fmt.Sprintf("%s %d/%d   /   %.1fs", locales.Uploaded(), s.Uploaded(), len(s.Puzzle.Targets), float64(s.ElapsedMS)/1000)).FontSize(22)
			ui.Text(c, reasonName(s.Reason)).TextColor(muted)
			if s.Mode == game.Tutorial {
				ui.Text(c, locales.TrainingDone()).TextColor(cyan)
			}
		})
		ui.Column(c).Grow(1).Gap(18).Justify(ui.Center).Children(func() {
			for _, item := range []struct {
				label string
				value int
			}{{locales.BaseScore(), score.Base}, {locales.ComboScore(), score.Combo}, {locales.CompleteScore(), score.Complete}, {locales.TimeScore(), score.Time}, {locales.BufferScore(), score.Buffer}} {
				ui.Row(c).Gap(10).Children(func() {
					ui.Text(c, item.label)
					ui.Spacer(c)
					ui.Text(c, fmt.Sprintf("+ %d", item.value)).TextColor(yellow)
				})
			}
			ui.Text(c, fmt.Sprintf("× %.1f", float64(s.Puzzle.Difficulty.Rules().Multiplier)/2)).FontSize(30).TextColor(red)
			ui.Text(c, fmt.Sprintf("%s %d   /   %s %d / %d", locales.Resumes(), s.Resumes, locales.Assistance(), s.Hints, s.Undos)).TextColor(muted)
		})
	})
	ui.Row(c).Gap(12).Children(func() {
		primary(c, locales.NewGame(), func() { w.navigate(setupPage) })
		button(c, locales.Replay(), func() { w.openReplay(s) })
		button(c, locales.Menu(), func() { w.navigate(menuPage) })
	})
}
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
func cycle(c *ui.Context, options []string, value *int) {
	if ui.Button(c, options[*value]).Padding(8, 14).Clicked() {
		*value = (*value + 1) % len(options)
	}
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
func (w *window) settingsView(c *ui.Context) {
	title(c, locales.Settings())
	ui.Scroll(c).Grow(1).MinHeight(0).TrackScroll(&w.settingsScroll).Children(func() {
		ui.Row(c).Gap(40).PaddingY(20).Children(func() {
			ui.Column(c).WidthPercent(45).Gap(16).Children(func() {
				sectionLabel(c, locales.Audio())
				for _, v := range []struct {
					label string
					value *float64
				}{{locales.MasterVolume(), &w.draft.Master}, {locales.MusicVolume(), &w.draft.Music}, {locales.EffectsVolume(), &w.draft.Effects}} {
					ui.Text(c, fmt.Sprintf("%s  %.0f%%", v.label, *v.value))
					ui.Slider(c, v.value, 0, 100).Label(v.label)
				}
				ui.Checkbox(c, &w.draft.Muted, locales.Muted())
				ui.Checkbox(c, &w.draft.Dynamic, locales.DynamicMusic())
				ui.Checkbox(c, &w.draft.BackgroundMute, locales.BackgroundMute())
				button(c, locales.PreviewAudio(), func() { w.previewAudio = !w.previewAudio; w.updateAudio(); w.audio.Effect("upload") })
			})
			ui.Column(c).Grow(1).Gap(18).Children(func() {
				sectionLabel(c, locales.Visuals())
				ui.Text(c, locales.Glitch())
				cycle(c, []string{locales.Off(), locales.Low(), locales.Medium(), locales.High()}, &w.draft.Glitch)
				ui.Checkbox(c, &w.draft.Reduced, locales.ReducedMotion())
				ui.Checkbox(c, &w.draft.LowTimePulse, locales.LowTimePulse())
				ui.Text(c, locales.TextScale())
				button(c, fmt.Sprintf("%d%%", w.draft.Scale), func() {
					w.draft.Scale += 25
					if w.draft.Scale > 150 {
						w.draft.Scale = 100
					}
				})
				ui.Text(c, locales.Language())
				ui.Row(c).Gap(6).Wrap().Children(func() {
					for _, tag := range locales.Catalog.Languages() {
						b := ui.Button(c, i18n.Autonym(tag)).Lang(tag)
						if w.draft.Language == tag {
							b.TextColor(cyan).Border(1, cyan)
						}
						if b.Clicked() {
							w.draft.Language = tag
						}
					}
				})
			})
		})
	})
	ui.Row(c).Gap(12).Children(func() {
		primary(c, locales.Apply(), func() { w.leaveSettings(true) })
		button(c, locales.Cancel(), func() { w.leaveSettings(false) })
	})
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
func (w *window) drawGlitch(p *ui.Painter, r ui.Rect) {
	s := w.data.Settings
	if w.page == settingsPage {
		s = w.draft
	}
	if s.Reduced {
		return
	}
	if s.LowTimePulse && w.page == gamePage && w.session != nil && !w.session.Unlimited && w.session.Remaining() <= 10000 {
		alpha := float32(0.15 + 0.08*math.Sin(float64(time.Now().UnixMilli())/600))
		p.Stroke(r, red.Alpha(alpha), 0, 3)
	}
	if s.Glitch == 0 || time.Since(w.glitchStart) > 180*time.Millisecond {
		return
	}
	frame := int(time.Since(w.glitchStart).Milliseconds() / 40)
	colors := []ui.Color{red, cyan, yellow, ui.Hex("#B66CF5")}
	for i := 0; i < 12; i++ {
		x := r.X + float32((i*97+frame*29)%int(max(r.W-100, 1)))
		y := r.Y + 8 + float32(i%3)*5
		if i%2 == 0 {
			y = r.Y + r.H - 25 + float32(i%3)*5
		}
		p.Fill(ui.Rect{X: x, Y: y, W: float32(8 + (i*11)%72), H: float32(1 + s.Glitch)}, colors[(i+frame)%4].Alpha(0.65), 0)
	}
}
