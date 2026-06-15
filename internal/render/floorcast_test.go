package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestFloorCeilingDiffersFromFlatFill(t *testing.T) {
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	tex := defaultTextures()
	cam := newCamera(g.Player.Angle, cfg.FOV)

	flat := make([]byte, cfg.Width*cfg.Height*4)
	clearBackground(flat, cfg)

	cast := make([]byte, cfg.Width*cfg.Height*4)
	drawFloorCeiling(cast, g, cam, cfg, tex)

	diff := 0
	for i := range flat {
		if flat[i] != cast[i] {
			diff++
		}
	}
	if diff == 0 {
		t.Error("textured floor/ceiling produced the same pixels as the flat fill")
	}
}

func TestFloorCeilingFillsEveryRow(t *testing.T) {
	cfg := Config{Width: 64, Height: 80, FOV: 1.152}
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	cam := newCamera(g.Player.Angle, cfg.FOV)
	fb := make([]byte, cfg.Width*cfg.Height*4)
	drawFloorCeiling(fb, g, cam, cfg, defaultTextures())

	// Every row must be painted (alpha set), so there is no seam at the horizon.
	for y := range cfg.Height {
		i := (y*cfg.Width + cfg.Width/2) * 4
		if fb[i+3] != 255 {
			t.Fatalf("row %d not painted by floor/ceiling cast", y)
		}
	}
}

func TestFloorCeilingFallsBackWhenTexturesMissing(t *testing.T) {
	cfg := Config{Width: 32, Height: 32, FOV: 1.152}
	l, _ := world.Generate(world.Config{Width: 24, Height: 18, Seed: 1})
	g := sim.New(l)
	cam := newCamera(g.Player.Angle, cfg.FOV)
	fb := make([]byte, cfg.Width*cfg.Height*4)

	drawFloorCeiling(fb, g, cam, cfg, &textureSet{}) // no floor/ceiling textures
	// Should have fallen back to the flat fill without panicking.
	if fb[3] != 255 {
		t.Error("fallback flat fill did not paint the frame")
	}
}
