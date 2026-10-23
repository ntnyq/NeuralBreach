package game

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestGenerationProperties(t *testing.T) {
	for d := Difficulty(0); d < 4; d++ {
		for seed := uint64(0); seed < 250; seed++ {
			p, err := Generate(context.Background(), seed, d)
			if err != nil {
				t.Fatal(d, seed, err)
			}
			if err = p.Validate(); err != nil {
				t.Fatal(err)
			}
			r := d.Rules()
			if MinimumCodes(p.Targets) < r.MinPath || len(p.Solution) > r.MaxPath {
				t.Fatal("difficulty outside bounds")
			}
			again, _ := Generate(context.Background(), seed, d)
			if !reflect.DeepEqual(p, again) {
				t.Fatal("unstable generator")
			}
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Generate(ctx, 1, 0); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
