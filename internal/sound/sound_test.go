package sound

import (
	"math"
	"testing"
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
