package render

import (
	"image/color"
	"math"
	"sort"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

// spriteScale and fireballScale control billboard size relative to a wall at the
// same distance (1 ≈ wall height).
const (
	spriteScale   = 0.9
	fireballScale = 0.45
	itemScale     = 0.4
)

// billboard is a depth-sortable sprite (a demon, a projectile or an item).
type billboard struct {
	pos    sim.Vec2
	z      float64 // world height: the base for grounded sprites, else the centre
	tex    *texture
	scale  float64
	ground bool // anchor the sprite's base at z rather than centring on it
}

// drawSprites projects demons and projectiles into the view, sorts them
// far-to-near, and draws them after the walls, hiding columns that fall behind
// nearer geometry using the wall depth buffer.
func drawSprites(fb []byte, zbuf []float64, g *sim.Game, cam camera, cfg Config, tx *textureSet) {
	w, h := cfg.Width, cfg.Height
	px, py := g.Player.Pos.X, g.Player.Pos.Y

	items := make([]billboard, 0, len(g.Entities)+len(g.Projectiles))
	for _, e := range g.Entities {
		items = append(items, billboard{pos: e.Pos, z: e.Z, tex: demonTexture(tx, e), scale: spriteScale, ground: true})
	}
	for _, p := range g.Projectiles {
		if p.Alive {
			items = append(items, billboard{pos: p.Pos, z: p.Z, tex: tx.fireball, scale: fireballScale})
		}
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

	// Order by descending distance so nearer sprites overdraw farther ones.
	sort.Slice(items, func(a, b int) bool {
		da := (items[a].pos.X-px)*(items[a].pos.X-px) + (items[a].pos.Y-py)*(items[a].pos.Y-py)
		db := (items[b].pos.X-px)*(items[b].pos.X-px) + (items[b].pos.Y-py)*(items[b].pos.Y-py)
		return da > db
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
		drawBillboard(fb, zbuf, cfg, screenX, top, size, depth, it.tex)
	}
}

// demonTexture picks the frame for a demon's variant and state.
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

// drawBillboard renders one textured sprite centred at screenX with its top at
// the given row, skipping transparent texels and columns occluded by nearer
// walls (via the depth buffer).
func drawBillboard(fb []byte, zbuf []float64, cfg Config, screenX, top, size int, depth float64, tex *texture) {
	w, h := cfg.Width, cfg.Height
	left := screenX - size/2

	for x := left; x < left+size; x++ {
		if x < 0 || x >= w {
			continue
		}
		if depth >= zbuf[x] {
			continue // hidden behind a nearer wall column
		}
		texX := int(float64(x-left) / float64(size) * float64(tex.w))
		for y := max(top, 0); y < min(top+size, h); y++ {
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
	f := math.Max(0.1, math.Min(1, 1.0/(1.0+depth*0.18)))
	return color.RGBA{
		R: uint8(float64(base.R) * f),
		G: uint8(float64(base.G) * f),
		B: uint8(float64(base.B) * f),
		A: 255,
	}
}
