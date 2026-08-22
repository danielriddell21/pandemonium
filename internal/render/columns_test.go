package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// boxLevel hand-builds a walled arena with a flat base floor and generous
// ceiling, for sculpting exact test terrain.
func boxLevel(w, h int) *world.Level {
	l := &world.Level{
		Width:  w,
		Height: h,
		Tiles:  make([]world.TileType, w*h),
		FloorH: make([]float64, w*h),
		CeilH:  make([]float64, w*h),
		Spawn:  world.Coord{X: 2, Y: h / 2},
		Exit:   world.Coord{X: w - 2, Y: h - 2},
	}
	for i := range l.CeilH {
		l.CeilH[i] = 1.5
	}
	for y := range h {
		for x := range w {
			if x == 0 || y == 0 || x == w-1 || y == h-1 {
				l.Tiles[y*w+x] = world.TileWall
			}
		}
	}
	return l
}

// sceneFor renders the level geometry from its spawn, facing +X.
func sceneFor(l *world.Level, cfg Config) []byte {
	g := sim.New(l)
	g.Entities = nil
	g.Items = nil
	g.Player.Angle = 0
	fb := make([]byte, cfg.Width*cfg.Height*4)
	zbuf := make([]float64, cfg.Width)
	loZ := make([]float64, cfg.Width)
	loH := make([]float64, cfg.Width)
	loRow := make([]int, cfg.Width)
	drawScene(fb, zbuf, loZ, loH, loRow, g, newCamera(0, cfg.FOV), cfg, defaultTextures(), 1)
	return fb
}

func TestStepFaceChangesTheView(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	flat := boxLevel(12, 8)
	raised := boxLevel(12, 8)
	for y := 1; y < 7; y++ {
		for x := 6; x < 11; x++ {
			raised.FloorH[y*12+x] = 0.5
		}
	}

	a, b := sceneFor(flat, cfg), sceneFor(raised, cfg)
	diff := 0
	for i := range a {
		if a[i] != b[i] {
			diff++
		}
	}
	if diff == 0 {
		t.Fatal("a raised floor changed nothing in the rendered view")
	}
	// The face must appear below the horizon, where flat ground showed floor.
	mid := cfg.Height/2 + 10
	rowDiff := 0
	for x := range cfg.Width {
		i := (mid*cfg.Width + x) * 4
		if a[i] != b[i] || a[i+1] != b[i+1] || a[i+2] != b[i+2] {
			rowDiff++
		}
	}
	if rowDiff == 0 {
		t.Error("no step face visible below the horizon")
	}
}

func TestLoweredCeilingChangesTheView(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	flat := boxLevel(12, 8)
	dropped := boxLevel(12, 8)
	for y := 1; y < 7; y++ {
		for x := 6; x < 11; x++ {
			dropped.CeilH[y*12+x] = 0.9
		}
	}

	a, b := sceneFor(flat, cfg), sceneFor(dropped, cfg)
	// The upper face must appear above the horizon.
	mid := cfg.Height/2 - 10
	rowDiff := 0
	for x := range cfg.Width {
		i := (mid*cfg.Width + x) * 4
		if a[i] != b[i] || a[i+1] != b[i+1] || a[i+2] != b[i+2] {
			rowDiff++
		}
	}
	if rowDiff == 0 {
		t.Error("no ceiling face visible above the horizon")
	}
}

func TestSpriteStandsOnItsFloor(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	l := boxLevel(12, 8)
	render := func(z float64) []byte {
		g := sim.New(l)
		g.Items = nil
		g.Player.Angle = 0
		g.Entities = []sim.Entity{{Pos: sim.Vec2{X: 6.5, Y: 4.5}, Z: z, State: sim.Active, Health: 60, Alive: true}}
		fb := make([]byte, cfg.Width*cfg.Height*4)
		zbuf := make([]float64, cfg.Width)
		loZ := make([]float64, cfg.Width)
		loH := make([]float64, cfg.Width)
		loRow := make([]int, cfg.Width)
		cam := newCamera(0, cfg.FOV)
		drawScene(fb, zbuf, loZ, loH, loRow, g, cam, cfg, defaultTextures(), 1)
		drawSprites(fb, zbuf, loZ, loH, loRow, g, cam, cfg, defaultTextures())
		return fb
	}
	topMost := func(fb, base []byte) int {
		for y := range cfg.Height {
			for x := range cfg.Width {
				i := (y*cfg.Width + x) * 4
				if fb[i] != base[i] || fb[i+1] != base[i+1] || fb[i+2] != base[i+2] {
					return y
				}
			}
		}
		return cfg.Height
	}
	base := sceneFor(l, cfg)
	ground := topMost(render(0), base)
	raisedTop := topMost(render(0.5), base)
	if raisedTop >= ground {
		t.Errorf("sprite raised to z=0.5 should draw higher on screen: top %d vs %d", raisedTop, ground)
	}
}
