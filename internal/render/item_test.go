package render

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestItemBillboardIsDrawn(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	g.Entities = nil
	g.Projectiles = nil

	// Find a facing where a tile ~1.6 ahead is open floor, and put an item there.
	p := g.Player.Pos
	placed := false
	for _, ang := range []float64{0, math.Pi / 2, math.Pi, -math.Pi / 2} {
		spot := sim.Vec2{X: p.X + 1.6*math.Cos(ang), Y: p.Y + 1.6*math.Sin(ang)}
		if g.World.Solid(int(spot.X), int(spot.Y)) {
			continue
		}
		g.Player.Angle = ang
		g.Items = []sim.ItemState{{Kind: world.ItemHealth, Pos: spot}}
		placed = true
		break
	}
	if !placed {
		t.Skip("no open tile ahead for item placement")
	}

	tex := defaultTextures()
	cam := testCam(g, g.Player.Angle, cfg.FOV)

	// Render the scene, then draw sprites onto a copy: the item must change pixels.
	base := make([]byte, cfg.Width*cfg.Height*4)
	zb := make([]float64, cfg.Width)
	loZ := make([]float64, cfg.Width)
	loH := make([]float64, cfg.Width)
	loRow := make([]int, cfg.Width)
	drawScene(base, zb, loZ, loH, loRow, g, cam, cfg, tex, 1)

	withItem := make([]byte, len(base))
	copy(withItem, base)
	drawSprites(withItem, zb, loZ, loH, loRow, g, cam, cfg, tex)

	changed := 0
	for i := range base {
		if base[i] != withItem[i] {
			changed++
		}
	}
	if changed == 0 {
		t.Error("item billboard changed no pixels in the view")
	}
}
