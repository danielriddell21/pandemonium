package render

import (
	"image/color"

	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/paint"
	"github.com/danielriddell21/crucible/raycast"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

const (
	spriteScale   = 0.9
	fireballScale = 0.45
	itemScale     = 0.4
)

type billboard struct {
	pos    sim.Vec2
	z      float64
	tex    *texture
	scale  float64
	ground bool
}

func drawSprites(fb []byte, zbuf, loZ, loH []float64, loRow []int, g *sim.Game, cam raycast.Camera, cfg Config, tx *textureSet) {
	w, h := cfg.Width, cfg.Height
	px, py := g.Player.Pos.X, g.Player.Pos.Y

	items := collectBillboards(g, tx)

	// Order by descending distance so nearer sprites overdraw farther ones.
	raycast.SortFarToNear(items, func(b billboard) float64 {
		dx, dy := b.pos.X-px, b.pos.Y-py
		return dx*dx + dy*dy
	})

	// Each sprite is anchored on its own world height (grounded sprites stand
	// on it, floating projectiles are centred on it), so it projects through
	// the engine's height-aware ProjectAt.
	eyeZ := g.EyeZ()
	for _, it := range items {
		pl, ok := cam.ProjectAt(geom.Vec2{X: it.pos.X, Y: it.pos.Y}, it.z, eyeZ, w, h, it.scale, it.ground)
		if !ok {
			continue
		}
		drawBillboard(fb, zbuf, loZ, loH, loRow, cfg, pl.ScreenX, pl.Top, pl.Size, pl.Depth, it.z, it.tex)
	}
}

func collectBillboards(g *sim.Game, tx *textureSet) []billboard {
	items := make([]billboard, 0, len(g.Entities)+len(g.Projectiles))
	for _, e := range g.Entities {
		if e.Kind == sim.Barrel {
			switch e.State {
			case sim.Dead:
				continue // burst and gone
			case sim.Dying:
				items = append(items, billboard{pos: e.Pos, z: e.Z, tex: tx.fireball, scale: spriteScale, ground: true})
			default:
				items = append(items, billboard{pos: e.Pos, z: e.Z, tex: tx.barrel, scale: spriteScale * 0.7, ground: true})
			}
			continue
		}
		items = append(items, billboard{pos: e.Pos, z: e.Z, tex: demonTexture(tx, e), scale: spriteScale, ground: true})
	}
	for _, p := range g.Projectiles {
		if !p.Alive {
			continue
		}
		tex := tx.fireball
		if p.Splash {
			tex = tx.rocket // the player's rocket reads differently from a fireball
		}
		items = append(items, billboard{pos: p.Pos, z: p.Z, tex: tex, scale: fireballScale})
	}
	for _, it := range g.Items {
		if it.Taken {
			continue
		}
		if t := tx.itemTexture(it.Kind); t != nil {
			z := g.World.FloorAt(int(it.Pos.X), int(it.Pos.Y))
			items = append(items, billboard{pos: it.Pos, z: z, tex: t, scale: itemScale, ground: true})
		}
	}
	return items
}

func demonTexture(tx *textureSet, e sim.Entity) *texture {
	art := tx.demon[e.Sprite%len(tx.demon)]
	switch e.State {
	case sim.Dead:
		return art.dead[len(art.dead)-1]
	case sim.Dying:
		i := e.Frame
		if i >= len(art.dead) {
			i = len(art.dead) - 1
		}
		return art.dead[i]
	default:
		return art.walk[e.Frame%len(art.walk)]
	}
}

func drawBillboard(fb []byte, zbuf, loZ, loH []float64, loRow []int, cfg Config, screenX, top, size int, depth, baseZ float64, tex *texture) {
	w, h := cfg.Width, cfg.Height
	left := screenX - size/2

	for x := left; x < left+size; x++ {
		if x < 0 || x >= w {
			continue
		}
		if depth >= zbuf[x] {
			continue // hidden behind a nearer wall column
		}
		// Beyond a near lip (a low wall, stair, lift or ledge), a sprite that sits
		// in the lower area behind it is blocked below the lip's top edge — but one
		// standing at or above the lip is not.
		behindLip := depth >= loZ[x] && baseZ < loH[x]
		texX := int(float64(x-left) / float64(size) * float64(tex.w))
		for y := max(top, 0); y < min(top+size, h); y++ {
			if behindLip && y > loRow[x] {
				continue
			}
			texY := int(float64(y-top) / float64(size) * float64(tex.h))
			texel := tex.at(texX, texY)
			if texel.A < 128 {
				continue // transparent
			}
			setPixel(fb, w, x, y, shadeRGBA(texel, depth))
		}
	}
}

func shadeRGBA(base color.RGBA, depth float64) color.RGBA {
	f := max(0.1, min(1, 1.0/(1.0+depth*0.18)))
	return paint.Scale(base, f)
}
