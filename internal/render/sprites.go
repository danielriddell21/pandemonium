package render

import (
	"cmp"
	"image/color"
	"slices"

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

func drawSprites(fb []byte, zbuf, loZ, loH []float64, loRow []int, g *sim.Game, cam camera, cfg Config, tx *textureSet) {
	w, h := cfg.Width, cfg.Height
	px, py := g.Player.Pos.X, g.Player.Pos.Y

	items := collectBillboards(g, tx)

	// Order by descending distance so nearer sprites overdraw farther ones.
	slices.SortFunc(items, func(a, b billboard) int {
		da := (a.pos.X-px)*(a.pos.X-px) + (a.pos.Y-py)*(a.pos.Y-py)
		db := (b.pos.X-px)*(b.pos.X-px) + (b.pos.Y-py)*(b.pos.Y-py)
		return cmp.Compare(db, da)
	})

	// Inverse of the [plane | dir] matrix maps world offsets into camera space.
	invDet := 1.0 / (cam.planeX*cam.dirY - cam.dirX*cam.planeY)
	eyeZ := g.EyeZ()

	for _, it := range items {
		relX, relY := it.pos.X-px, it.pos.Y-py
		transformX := invDet * (cam.dirY*relX - cam.dirX*relY)
		depth := invDet * (-cam.planeY*relX + cam.planeX*relY)
		if depth <= 0.01 {
			continue // behind the camera
		}
		screenX := int(float64(w) / 2 * (1 + transformX/depth))
		size := int(float64(h) / depth * it.scale)
		if size <= 0 {
			continue
		}
		// Project the sprite's world height: grounded sprites stand on it,
		// floating ones (projectiles) are centred on it.
		anchor := int(float64(h)/2 + (eyeZ-it.z)*float64(h)/depth)
		top := anchor - size
		if !it.ground {
			top = anchor - size/2
		}
		drawBillboard(fb, zbuf, loZ, loH, loRow, cfg, screenX, top, size, depth, it.z, it.tex)
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
	return color.RGBA{
		R: uint8(float64(base.R) * f),
		G: uint8(float64(base.G) * f),
		B: uint8(float64(base.B) * f),
		A: 255,
	}
}
