package render

import (
	"testing"

	"github.com/danielriddell21/crucible/level"
	"github.com/danielriddell21/crucible/paint"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func skyBox(sky bool) *world.Level {
	const w, h = 10, 9
	base := level.New(w, h, 0)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			base.Set(x, y, world.TileFloor)
			if sky {
				base.Sky[y*w+x] = true
			}
		}
	}
	base.Spawn = world.Coord{X: 2, Y: 4}
	base.Exit = world.Coord{X: 8, Y: 4}
	return &world.Level{Level: base}
}

func renderBox(l *world.Level) []byte {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	g := sim.New(l)
	g.Entities = nil
	g.Player.Pos = sim.Vec2{X: 1.5, Y: 4.5}
	g.Player.Angle = 0 // look across the room toward the far wall
	fb := make([]byte, cfg.Width*cfg.Height*4)
	zb := make([]float64, cfg.Width)
	loZ := make([]float64, cfg.Width)
	loH := make([]float64, cfg.Width)
	loRow := make([]int, cfg.Width)
	drawScene(fb, zb, loZ, loH, loRow, g, testCam(g, 0, cfg.FOV), cfg, defaultTextures(), 1)
	return fb
}

func TestSkyChangesTheCeiling(t *testing.T) {
	stone := renderBox(skyBox(false))
	sky := renderBox(skyBox(true))

	diff := 0
	for i := range stone {
		if stone[i] != sky[i] {
			diff++
		}
	}
	if diff == 0 {
		t.Fatal("opening the room to the sky changed nothing in the view")
	}
	if !skyBox(true).SkyAt(5, 4) {
		t.Error("expected the open room to be flagged as sky")
	}
}

func TestSkyGloomDarkens(t *testing.T) {
	// The sky takes the run's gloom like everything else: a sky pixel must be
	// brighter at full light than under heavy gloom.
	bright := lerpColor(palette.skyTop, palette.skyHorizon, 0.5)
	dim := paint.Scale(bright, 0.4)
	if int(dim.R)+int(dim.G)+int(dim.B) >= int(bright.R)+int(bright.G)+int(bright.B) {
		t.Error("gloom should darken the sky")
	}
}
