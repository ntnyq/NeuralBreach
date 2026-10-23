package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/internal/game"
	"github.com/ntnyq/cyber-game/locales"
)

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

func (w *window) drawBackground(p *ui.Painter, r ui.Rect) {
	if w.page != menuPage || menuBackground == nil {
		return
	}
	p.Image(menuBackground, r, ui.Cover)
	p.FillGradient(r, ui.LinearGradient{Angle: 90, Stops: []ui.GradientStop{{Color: black.Alpha(0.45), At: 0}, {Color: black.Alpha(0.2), At: 0.42}, {Color: black.Alpha(0.04), At: 1}}}, 0)
}
