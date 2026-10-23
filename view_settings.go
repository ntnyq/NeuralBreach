package main

import (
	"fmt"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/ntnyq/cyber-game/locales"
)

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
