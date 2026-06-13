package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// The renderer draws into a plain RGBA byte buffer (4 bytes per pixel, row
// major) which the Ebiten layer uploads in one call. Keeping the drawing here as
// pure pixel writes leaves this code free of any graphics dependency.

// clearBackground fills the ceiling (top half) and floor (bottom half).
func clearBackground(fb []byte, cfg Config) {
	mid := cfg.Height / 2
	fillRows(fb, cfg.Width, 0, mid, palette.ceiling)
	fillRows(fb, cfg.Width, mid, cfg.Height, palette.floor)
}

// drawWalls casts one ray per screen column, draws the texture-mapped, distance-
// shaded wall slice, and records each column's perpendicular distance into zbuf
// for sprite occlusion.
func drawWalls(fb []byte, zbuf []float64, g *sim.Game, cam camera, cfg Config, tx *textureSet) {
	w, h := cfg.Width, cfg.Height
	px, py := g.Player.Pos.X, g.Player.Pos.Y

	for x := range w {
		dx, dy := cam.rayDir(x, w)
		ht := castRay(g.World, px, py, dx, dy)
		zbuf[x] = ht.perpDist

		lineH := int(float64(h) / ht.perpDist)
		if lineH < 1 {
			lineH = 1
		}
		y0 := h/2 - lineH/2
		y1 := h/2 + lineH/2

		tex := tx.wall
		if g.World.Level.At(ht.mapX, ht.mapY) == world.TileDoor {
			tex = tx.door
		}

		// Texture column, flipped so the image faces the camera consistently.
		texX := int(ht.wallX * float64(tex.w))
		if texX >= tex.w {
			texX = tex.w - 1
		}
		if (ht.side == 0 && dx > 0) || (ht.side == 1 && dy < 0) {
			texX = tex.w - 1 - texX
		}

		start, end := y0, y1
		if start < 0 {
			start = 0
		}
		if end >= h {
			end = h - 1
		}
		step := float64(tex.h) / float64(lineH)
		texPos := float64(start-y0) * step
		for y := start; y <= end; y++ {
			texel := tex.at(texX, int(texPos))
			texPos += step
			setPixel(fb, w, x, y, shade(texel, ht.perpDist, ht.side))
		}
	}
}

// shade darkens a colour with distance (for the dim look) and a little extra for
// north/south faces so edges read clearly.
func shade(c color.RGBA, dist float64, side int) color.RGBA {
	f := 1.0 / (1.0 + dist*0.18)
	if side == 1 {
		f *= 0.72
	}
	f = math.Max(0.08, math.Min(1, f))
	return color.RGBA{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: 255,
	}
}

// fillRows fills whole rows [y0, y1) with a solid colour.
func fillRows(fb []byte, w, y0, y1 int, c color.RGBA) {
	for y := y0; y < y1; y++ {
		for x := range w {
			setPixel(fb, w, x, y, c)
		}
	}
}

// setPixel writes one opaque pixel; out-of-bounds writes are ignored.
func setPixel(fb []byte, w, x, y int, c color.RGBA) {
	i := (y*w + x) * 4
	if i < 0 || i+3 >= len(fb) {
		return
	}
	fb[i] = c.R
	fb[i+1] = c.G
	fb[i+2] = c.B
	fb[i+3] = 255
}
