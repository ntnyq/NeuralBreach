package sound

import (
	"math"
	"testing"

	"github.com/egoist/mygo/audio"
)

func TestSynthLevelsAndFade(t *testing.T) {
	s := &synth{}
	s.gain.Store(math.Float64bits(0.65))
	buf := make([]float32, 48000)
	for _, stage := range []Stage{Menu, Observe, Breach, Urgent, Result, Paused} {
		s.stage.Store(int32(stage))
		n, err := s.Read(buf)
		if err != nil || n != len(buf) {
			t.Fatal(n, err)
		}
		energy := 0.0
		for _, v := range buf {
			if math.IsNaN(float64(v)) || math.Abs(float64(v)) > 0.7 {
				t.Fatal("unbounded sample", v)
			}
			energy += float64(v * v)
		}
		if stage != Paused && energy == 0 {
			t.Fatal("silent music")
		}
	}
	s.Read(buf)
	if s.level > 0.001 {
		t.Fatal("pause did not fade", s.level)
	}
}

func TestMuteSilencesExistingEffectsAndRestoresVolume(t *testing.T) {
	previous := audio.Volume()
	t.Cleanup(func() { audio.SetVolume(previous) })
	// Paused memory players let us inspect mixer gains without opening a device.
	voice := audio.NewSound(24000, 1, []float32{1}).NewPlayer().SetVolume(0.85)
	t.Cleanup(voice.Close)
	e := &Engine{source: &synth{}, voices: []*audio.Player{voice}}
	config := Config{Master: 80, Music: 65, Effects: 85, Dynamic: true}
	e.Update(Result, config)
	if gain := audio.Volume() * voice.Volume(); math.Abs(gain-0.68) > 1e-9 {
		t.Fatalf("unexpected foreground effect gain %f", gain)
	}
	config.Muted = true
	e.Update(Paused, config)
	if gain := audio.Volume() * voice.Volume(); gain != 0 {
		t.Fatalf("muted effect remains audible at gain %f", gain)
	}
	config.Muted = false
	e.Update(Result, config)
	if gain := audio.Volume() * voice.Volume(); math.Abs(gain-0.68) > 1e-9 {
		t.Fatalf("effect volume was not restored: %f", gain)
	}
}

func TestMutedEffectsDoNotStartPlayers(t *testing.T) {
	previous := audio.Volume()
	t.Cleanup(func() { audio.SetVolume(previous) })
	played := 0
	e := &Engine{
		source:  &synth{},
		effects: map[string]*audio.Sound{"complete": audio.NewSound(24000, 1, []float32{1})},
		play: func(s *audio.Sound, volume float64) *audio.Player {
			played++
			p := s.NewPlayer().SetVolume(volume)
			t.Cleanup(p.Close)
			return p
		},
	}
	config := Config{Master: 80, Music: 65, Effects: 85, Muted: true}
	e.Update(Paused, config)
	e.Effect("complete")
	if played != 0 || len(e.voices) != 0 {
		t.Fatal("an effect started while muted")
	}
	config.Muted = false
	e.Update(Result, config)
	e.Effect("complete")
	if played != 1 || len(e.voices) != 1 || e.voices[0].Volume() != 0.85 {
		t.Fatal("foreground effect did not resume at the configured volume")
	}
}
