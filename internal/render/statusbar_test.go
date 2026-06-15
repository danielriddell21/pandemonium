package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestStatusBarDrawsPanel(t *testing.T) {
	cfg := Config{Width: 320, Height: 200, FOV: 1.152}
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 5})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	fb := make([]byte, cfg.Width*cfg.Height*4)

	drawStatusBar(fb, cfg, g, defaultTextures())

	// The bottom panel rows must be painted; rows above it must be untouched.
	bottomMid := ((cfg.Height-1)*cfg.Width + cfg.Width/2) * 4
	if fb[bottomMid+3] != 255 {
		t.Error("status bar panel not drawn at the bottom")
	}
	topMid := (10*cfg.Width + cfg.Width/2) * 4
	if fb[topMid+3] != 0 {
		t.Error("status bar painted above its panel region")
	}
}

func TestStatusBarReflectsKeysHeld(t *testing.T) {
	cfg := Config{Width: 320, Height: 200, FOV: 1.152}
	l, _ := world.Generate(world.Config{Width: 32, Height: 24, Seed: 5})
	tx := defaultTextures()

	none := sim.New(l)
	withKey := sim.New(l)
	withKey.Player.Keys = map[world.ItemKind]bool{world.ItemKeyRed: true}

	a := make([]byte, cfg.Width*cfg.Height*4)
	b := make([]byte, cfg.Width*cfg.Height*4)
	drawStatusBar(a, cfg, none, tx)
	drawStatusBar(b, cfg, withKey, tx)

	// A held key lights its pip, so the two bars must differ.
	same := true
	for i := range a {
		if a[i] != b[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("holding a key did not change the status bar")
	}
}

func TestFaceBandTracksHealth(t *testing.T) {
	cases := []struct {
		frac float64
		want int
	}{{1, 0}, {0.5, 1}, {0.2, 2}, {0, 3}}
	for _, c := range cases {
		if got := faceBand(c.frac); got != c.want {
			t.Errorf("faceBand(%v) = %d, want %d", c.frac, got, c.want)
		}
	}
}
