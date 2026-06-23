package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestGloomDarkensTheFrame(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	g.Entities = nil

	r := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	bright := brightnessSum(append([]byte(nil), r.Frame(g)...))
	r.SetGloom(0.4)
	dim := brightnessSum(r.Frame(g))

	if dim >= bright {
		t.Errorf("gloom should darken the frame: bright=%d dim=%d", bright, dim)
	}
}

func TestGloomClamped(t *testing.T) {
	r := NewRenderer(DefaultConfig())
	r.SetGloom(-5)
	if r.gloom < 0.3 {
		t.Errorf("gloom should clamp at 0.3, got %v", r.gloom)
	}
	r.SetGloom(99)
	if r.gloom > 1 {
		t.Errorf("gloom should clamp at 1, got %v", r.gloom)
	}
}

func brightnessSum(fb []byte) int {
	sum := 0
	for i := 0; i+3 < len(fb); i += 4 {
		sum += int(fb[i]) + int(fb[i+1]) + int(fb[i+2])
	}
	return sum
}
