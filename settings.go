package main

import (
	"github.com/egoist/mygo"
	"github.com/ntnyq/cyber-game/internal/sound"
)

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
	w.audio.Update(stage, sound.Config{
		Master:  settings.Master,
		Music:   settings.Music,
		Effects: settings.Effects,
		Muted:   settings.Muted,
		Dynamic: settings.Dynamic,
	})
}
