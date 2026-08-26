package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestAutomapDrawsInCorner(t *testing.T) {
	cfg := Config{Width: 320, Height: 200, FOV: 1.152}
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 5})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	for range 300 { // explore so the map has something to show
		g.Tick(sim.Input{Forward: 1}, 1.0/60.0)
	}
	fb := make([]byte, cfg.Width*cfg.Height*4)
	drawAutomap(fb, cfg, g)

	// Something is drawn in the top-right quadrant.
	tr := (20*cfg.Width + cfg.Width - 20) * 4
	if fb[tr+3] != 255 {
		t.Error("automap did not draw in the top-right corner")
	}
	// Nothing drawn in the bottom-left (map is corner-only).
	bl := ((cfg.Height - 20) * cfg.Width) * 4
	if fb[bl+3] != 0 {
		t.Error("automap painted outside its corner region")
	}
}

func TestAutomapToggle(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	l, _ := world.Generate(world.Config{Width: 32, Height: 24, Seed: 5})
	g := sim.New(l)
	r := NewRenderer(cfg)

	off := append([]byte(nil), r.Frame(g)...)
	r.SetAutomap(true)
	on := r.Frame(g)

	diff := false
	for i := range off {
		if off[i] != on[i] {
			diff = true
			break
		}
	}
	if !diff {
		t.Error("enabling the automap did not change the frame")
	}
}
