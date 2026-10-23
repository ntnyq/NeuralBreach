package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/locales"
)

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
