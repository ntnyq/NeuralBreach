package main

import (
	"testing"

	"github.com/ntnyq/cyber-game/internal/sound"
)

func TestBackgroundAudioMuteAndRestore(t *testing.T) {
	for _, p := range []page{menuPage, gamePage, pausePage, resultPage, settingsPage} {
		w := fixtureWindow()
		w.page = p
		w.draft = w.data.Settings
		w.previewAudio = true
		foreground, original := w.audioState()
		w.background = true
		stage, muted := w.audioState()
		if stage != sound.Paused || !muted.Muted {
			t.Fatalf("page %d: background did not mute all audio: stage=%d config=%+v", p, stage, muted)
		}
		if w.data.Settings.Muted || w.draft.Muted || muted.Master != original.Master || muted.Effects != original.Effects {
			t.Fatalf("page %d: background mute changed the volume preferences", p)
		}
		w.background = false
		stage, restored := w.audioState()
		if stage != foreground || restored != original {
			t.Fatalf("page %d: focus did not restore audio: stage=%d config=%+v", p, stage, restored)
		}
		w.background = true
		w.data.Settings.BackgroundMute = false
		w.draft.BackgroundMute = false
		stage, unmuted := w.audioState()
		if stage != foreground || unmuted != original {
			t.Fatalf("page %d: disabled background mute still changed audio", p)
		}
		w.data.Settings.Muted = true
		w.draft.Muted = true
		w.background = false
		_, manual := w.audioState()
		if !manual.Muted {
			t.Fatalf("page %d: focus cleared manual mute", p)
		}
	}
}
