package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func skyBox(sky bool) *world.Level {
	const w, h = 10, 9
	l := &world.Level{
		Width: w, Height: h,
		Tiles:   make([]world.TileType, w*h),
		FloorH:  make([]float64, w*h),
		CeilH:   make([]float64, w*h),
		Light:   make([]float64, w*h),
		Theme:   make([]uint8, w*h),
		WallTop: make([]float64, w*h),
		Sky:     make([]bool, w*h),
		Spawn:   world.Coord{X: 2, Y: 4},
		Exit:    world.Coord{X: 8, Y: 4},
	}
	for i := range l.CeilH {
		l.CeilH[i] = 1
		l.Light[i] = 1
	}
	for y := range h {
		for x := range w {
			if x == 0 || y == 0 || x == w-1 || y == h-1 {
				l.Tiles[y*w+x] = world.TileWall
			} else if sky {
				l.Sky[y*w+x] = true
			}
		}
	}
	return l
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
	drawScene(fb, zb, loZ, loH, loRow, g, newCamera(0, cfg.FOV), cfg, defaultTextures(), 1)
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
	dim := scaleColor(bright, 0.4)
	if int(dim.R)+int(dim.G)+int(dim.B) >= int(bright.R)+int(bright.G)+int(bright.B) {
		t.Error("gloom should darken the sky")
	}
}
