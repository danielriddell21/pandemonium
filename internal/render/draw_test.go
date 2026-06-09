package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// renderFrame runs the full CPU draw pipeline into a fresh buffer, the same way
// the Ebiten layer does, so the renderer can be exercised without a GPU.
func renderFrame(t *testing.T, seed int64, cfg Config) ([]byte, []float64) {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: seed})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	fb := make([]byte, cfg.Width*cfg.Height*4)
	zbuf := make([]float64, cfg.Width)
	cam := newCamera(g.Player.Angle, cfg.FOV)
	clearBackground(fb, cfg)
	drawWalls(fb, zbuf, g, cam, cfg)
	drawSprites(fb, zbuf, g, cam, cfg)
	return fb, zbuf
}

func TestDrawWallsProducesGeometry(t *testing.T) {
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}
	fb, zbuf := renderFrame(t, 7, cfg)

	// The depth buffer must be filled with finite positive distances.
	for x, d := range zbuf {
		if d <= 0 {
			t.Fatalf("zbuf[%d] = %v, want positive", x, d)
		}
	}

	// Some pixels in the vertical middle band must be wall-coloured, i.e. differ
	// from both the flat ceiling and floor.
	wallPixels := 0
	for x := range cfg.Width {
		y := cfg.Height / 2
		i := (y*cfg.Width + x) * 4
		c := [3]byte{fb[i], fb[i+1], fb[i+2]}
		if c != [3]byte{palette.ceiling.R, palette.ceiling.G, palette.ceiling.B} &&
			c != [3]byte{palette.floor.R, palette.floor.G, palette.floor.B} {
			wallPixels++
		}
	}
	if wallPixels == 0 {
		t.Error("no wall pixels rendered across the middle scanline")
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
