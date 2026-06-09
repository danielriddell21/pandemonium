package render

import (
	"image/color"
	"math"
	"sort"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

// spriteScale controls how large billboards appear relative to a wall at the
// same distance (1 ≈ wall height).
const spriteScale = 0.9

// drawSprites projects each demon into the view, sorts them far-to-near, and
// draws them after the walls, hiding columns that fall behind nearer geometry
// using the wall depth buffer.
func drawSprites(fb []byte, zbuf []float64, g *sim.Game, cam camera, cfg Config) {
	w, h := cfg.Width, cfg.Height
	px, py := g.Player.Pos.X, g.Player.Pos.Y

	// Order by descending distance so nearer sprites overdraw farther ones.
	order := make([]int, 0, len(g.Entities))
	for i, e := range g.Entities {
		if e.Alive {
			order = append(order, i)
		}
	}
	sort.Slice(order, func(a, b int) bool {
		return distSq(g.Entities[order[a]], px, py) > distSq(g.Entities[order[b]], px, py)
	})

	// Inverse of the [plane | dir] matrix maps world offsets into camera space.
	invDet := 1.0 / (cam.planeX*cam.dirY - cam.dirX*cam.planeY)

	for _, idx := range order {
		e := g.Entities[idx]
		relX, relY := e.Pos.X-px, e.Pos.Y-py

		transformX := invDet * (cam.dirY*relX - cam.dirX*relY)
		depth := invDet * (-cam.planeY*relX + cam.planeX*relY)
		if depth <= 0.01 {
			continue // behind the camera
		}

		screenX := int(float64(w) / 2 * (1 + transformX/depth))
		size := int(float64(h) / depth * spriteScale)
		if size <= 0 {
			continue
		}

		drawBillboard(fb, zbuf, cfg, screenX, size, depth, palette.sprite[e.Sprite%len(palette.sprite)])
	}
}

// drawBillboard renders one sprite as a shaded blob centred at screenX, occluded
// per-column by the wall depth buffer.
func drawBillboard(fb []byte, zbuf []float64, cfg Config, screenX, size int, depth float64, base color.RGBA) {
	w, h := cfg.Width, cfg.Height

	top := h/2 - size/2
	bottom := top + size
	left := screenX - size/2
	right := left + size

	col := shadeRGBA(base, depth)
	rx, ry := float64(size)/2, float64(size)/2
	cx, cy := float64(screenX), float64(top)+ry

	for x := left; x < right; x++ {
		if x < 0 || x >= w {
			continue
		}
		if depth >= zbuf[x] {
			continue // hidden behind a nearer wall column
		}
		for y := max(top, 0); y < min(bottom, h); y++ {
			// Elliptical body so the demon isn't a hard rectangle.
			nx := (float64(x) - cx) / rx
			ny := (float64(y) - cy) / ry
			if nx*nx+ny*ny > 1 {
				continue
			}
			setPixel(fb, w, x, y, col)
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

func distSq(e sim.Entity, px, py float64) float64 {
	dx, dy := e.Pos.X-px, e.Pos.Y-py
	return dx*dx + dy*dy
}
