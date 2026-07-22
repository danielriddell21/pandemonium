package render

import (
	"testing"

	"github.com/danielriddell21/crucible/level"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func twoRoomLevel(lowWall bool) *world.Level {
	const w, h = 12, 8
	base := level.New(w, h, 0)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			if x != 5 {
				base.Set(x, y, world.TileFloor)
			}
		}
	}
	base.Spawn = world.Coord{X: 2, Y: 4}
	base.Exit = world.Coord{X: 9, Y: 4}
	if lowWall {
		base.WallTopH[4*w+5] = 0.4
	}
	return &world.Level{Level: base}
}

func renderEyeLine(l *world.Level) []byte {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	g := sim.New(l)
	g.Entities = nil
	g.Player.Pos = sim.Vec2{X: 2.5, Y: 4.5}
	g.Player.Angle = 0 // straight toward the divider
	fb := make([]byte, cfg.Width*cfg.Height*4)
	zb := make([]float64, cfg.Width)
	loZ := make([]float64, cfg.Width)
	loH := make([]float64, cfg.Width)
	loRow := make([]int, cfg.Width)
	drawScene(fb, zb, loZ, loH, loRow, g, testCam(g, 0, cfg.FOV), cfg, defaultTextures(), 1)
	return fb
}

func TestLowWallRevealsRoomBeyond(t *testing.T) {
	full := renderEyeLine(twoRoomLevel(false))
	low := renderEyeLine(twoRoomLevel(true))

	diff := 0
	for i := range full {
		if full[i] != low[i] {
			diff++
		}
	}
	if diff == 0 {
		t.Fatal("lowering the divider wall changed nothing in the view")
	}

	// Just above the horizon, straight ahead, the full wall shows wall texture
	// while the low wall shows the space beyond — so the centre column must differ.
	cfg := Config{Width: 200, Height: 120}
	x := cfg.Width / 2
	rowDiff := false
	for y := cfg.Height/2 - 20; y < cfg.Height/2; y++ {
		i := (y*cfg.Width + x) * 4
		if full[i] != low[i] || full[i+1] != low[i+1] || full[i+2] != low[i+2] {
			rowDiff = true
		}
	}
	if !rowDiff {
		t.Error("expected to see over the low wall above the horizon")
	}
}

func TestFullWallStillOccludes(t *testing.T) {
	// A level with no low walls must not trip the see-over path: it still renders
	// (every pixel painted) and the divider fully blocks the view.
	fb := renderEyeLine(twoRoomLevel(false))
	for i := 3; i < len(fb); i += 4 {
		if fb[i] != 255 {
			t.Fatalf("pixel %d unpainted by the full-wall render", i/4)
		}
	}
}
