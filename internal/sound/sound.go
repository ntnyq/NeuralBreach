// Package sound synthesizes original industrial-electronic music and effects.
// One decoder owns the shared sample clock; state changes crossfade its layers.
package sound

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo/audio"
)

// Config controls playback. Volumes are percentages in the range 0–100.
type Config struct {
	Master  float64
	Music   float64
	Effects float64
	Muted   bool
	Dynamic bool
}

type Stage int32

const (
	Menu Stage = iota
	Observe
	Breach
	Urgent
	Result
	Paused
)

type synth struct {
	stage                       atomic.Int32
	gain                        atomic.Uint64
	frame                       int64
	level, beatLevel, highLevel float64
}

func (*synth) SampleRate() int                 { return 24000 }
func (*synth) Channels() int                   { return 2 }
func (*synth) Len() int64                      { return -1 }
func (s *synth) SetPosition(frame int64) error { s.frame = frame; return nil }
func (s *synth) Read(buf []float32) (int, error) {
	stage := Stage(s.stage.Load())
	gain := math.Float64frombits(s.gain.Load())
	drums, high := 0.0, 0.0
	if stage == Breach || stage == Urgent {
		drums = 1
	}
	if stage == Urgent {
		high = 1
	}
	if stage == Paused {
		gain = 0
	}
	for i := 0; i+1 < len(buf); i += 2 {
		s.level += (gain - s.level) / 6000
		s.beatLevel += (drums - s.beatLevel) / 6000
		s.highLevel += (high - s.highLevel) / 6000
		t := float64(s.frame) / 24000
		beat := t / 0.6
		part := math.Mod(beat, 1)
		step := int(beat*2) % 16
		notes := [16]float64{55, 55, 65.406, 55, 49, 55, 73.416, 65.406, 55, 55, 82.407, 73.416, 49, 49, 65.406, 51.913}
		bass := math.Sin(2*math.Pi*notes[step]*t) * math.Exp(-math.Mod(beat*2, 1)*5)
		drone := (math.Sin(2*math.Pi*55*t) + 0.35*math.Sin(2*math.Pi*82.407*t)) * 0.10
		kick := math.Sin(2*math.Pi*(46*t+5*(1-math.Exp(-part*30)))) * math.Exp(-part*14)
		// A deterministic hash supplies quiet, high-passed metallic grit.
		x := uint64(s.frame)*6364136223846793005 + 1442695040888963407
		noise := float64(x>>40)/8388608 - 1
		hat := noise * math.Exp(-math.Mod(beat*4, 1)*24) * 0.10
		arp := math.Sin(2*math.Pi*notes[(step+4)%16]*8*t) * math.Exp(-math.Mod(beat*4, 1)*10)
		v := (drone + 0.12*bass + s.beatLevel*(0.16*kick+hat) + s.highLevel*0.045*arp) * s.level
		buf[i] = float32(math.Tanh(v))
		buf[i+1] = float32(math.Tanh(v + 0.012*math.Sin(2*math.Pi*110.15*t)*s.level))
		s.frame++
	}
	return len(buf) / 2 * 2, nil
}

type Engine struct {
	music     *audio.Player
	source    *synth
	effects   map[string]*audio.Sound
	voices    []*audio.Player
	play      func(*audio.Sound, float64) *audio.Player
	settings  Config
	lastError time.Time
}

func New() *Engine {
	s := &synth{}
	e := &Engine{source: s, effects: map[string]*audio.Sound{}, play: playSound}
	for name, freq := range map[string]float64{"select": 680, "move": 320, "invalid": 120, "upload": 1040, "complete": 880, "fail": 65, "glitch": 180} {
		duration := 0.12
		if name == "complete" {
			duration = 1.4
		}
		if name == "fail" {
			duration = 0.4
		}
		samples := make([]float32, int(24000*duration))
		for i := range samples {
			t := float64(i) / 24000
			f := freq
			if name == "complete" {
				f *= math.Pow(2, float64(int(t/0.22)%5)/12)
			}
			env := math.Min(t/0.008, 1) * math.Pow(1-t/duration, 2)
			v := math.Sin(2*math.Pi*f*t) + 0.2*math.Sin(2*math.Pi*f*2.01*t)
			samples[i] = float32(v * env * 0.16)
		}
		e.effects[name] = audio.NewSound(24000, 1, samples)
	}
	e.music = audio.NewPlayer(s)
	e.music.SetVolume(1).Play()
	return e
}
func (e *Engine) Update(stage Stage, s Config) {
	if e == nil {
		return
	}
	e.settings = s
	if !s.Dynamic && (stage == Breach || stage == Urgent) {
		stage = Observe
	}
	volume := s.Master / 100
	if s.Muted {
		volume = 0
	}
	audio.SetVolume(volume)
	e.source.stage.Store(int32(stage))
	e.source.gain.Store(math.Float64bits(s.Music / 100))
}
func (e *Engine) Effect(name string) {
	if e == nil || e.settings.Muted {
		return
	}
	if name == "invalid" && time.Since(e.lastError) < 150*time.Millisecond {
		return
	}
	if name == "invalid" {
		e.lastError = time.Now()
	}
	if s := e.effects[name]; s != nil {
		active := e.voices[:0]
		for _, p := range e.voices {
			if p.Playing() {
				active = append(active, p)
			} else {
				p.Close()
			}
		}
		e.voices = active
		p := e.play(s, e.settings.Effects/100)
		e.voices = append(e.voices, p)
	}
}

func playSound(s *audio.Sound, volume float64) *audio.Player {
	p := s.NewPlayer().SetVolume(volume)
	p.Play()
	return p
}

func (e *Engine) Close() {
	if e == nil {
		return
	}
	e.music.Close()
	for _, p := range e.voices {
		p.Close()
	}
}
func Error() error { return audio.Err() }
