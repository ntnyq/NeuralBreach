package main

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
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

func cycle(c *ui.Context, options []string, value *int) {
	if ui.Button(c, options[*value]).Padding(8, 14).Clicked() {
		*value = (*value + 1) % len(options)
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
