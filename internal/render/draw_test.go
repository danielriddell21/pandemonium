package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func renderFrame(t *testing.T, seed int64, cfg Config) ([]byte, []float64) {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: seed})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	tex := defaultTextures()
	fb := make([]byte, cfg.Width*cfg.Height*4)
	zbuf := make([]float64, cfg.Width)
	loZ := make([]float64, cfg.Width)
	loH := make([]float64, cfg.Width)
	loRow := make([]int, cfg.Width)
	cam := testCam(g, g.Player.Angle, cfg.FOV)
	drawScene(fb, zbuf, loZ, loH, loRow, g, cam, cfg, tex, 1)
	drawSprites(fb, zbuf, loZ, loH, loRow, g, cam, cfg, tex)
	return fb, zbuf
}

func TestDrawSceneProducesGeometry(t *testing.T) {
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	fb, zbuf := renderFrame(t, 7, cfg)

	// The depth buffer must be filled with finite positive distances.
	for x, d := range zbuf {
		if d <= 0 {
			t.Fatalf("zbuf[%d] = %v, want positive", x, d)
		}
	}

	// Every pixel must be painted: floors below, ceilings above, walls between.
	for i := 3; i < len(fb); i += 4 {
		if fb[i] != 255 {
			t.Fatalf("pixel %d left unpainted", i/4)
		}
	}
}

func TestDrawIsDeterministic(t *testing.T) {
	cfg := DefaultConfig()
	a, _ := renderFrame(t, 3, cfg)
	b, _ := renderFrame(t, 3, cfg)
	if len(a) != len(b) {
		t.Fatal("frame size mismatch")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("frames differ at byte %d", i)
		}
	}
}
