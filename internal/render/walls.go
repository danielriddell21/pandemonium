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

// drawWalls casts one ray per screen column, draws the shaded wall slice, and
// records each column's perpendicular distance into zbuf for sprite occlusion.
func drawWalls(fb []byte, zbuf []float64, g *sim.Game, cam camera, cfg Config) {
	w, h := cfg.Width, cfg.Height
	px, py := g.Player.Pos.X, g.Player.Pos.Y

	for x := range w {
		dx, dy := cam.rayDir(x, w)
		ht := castRay(g.World, px, py, dx, dy)
		zbuf[x] = ht.perpDist

		lineH := int(float64(h) / ht.perpDist)
		y0 := h/2 - lineH/2
		y1 := h/2 + lineH/2

		base := wallColor(g, ht)
		col := shade(base, ht.perpDist, ht.side)
		fillColumn(fb, w, h, x, y0, y1, col)
	}
}

// wallColor picks the flat colour for the struck wall cell: doors get their own
// tint, other walls are tinted by which face was hit.
func wallColor(g *sim.Game, ht hit) color.RGBA {
	if g.World.Level.At(ht.mapX, ht.mapY) == world.TileDoor {
		return palette.door
	}
	if ht.side == 0 {
		return palette.wallX
	}
	return palette.wallY
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

// fillColumn fills a vertical span [y0, y1] at column x, clipped to the screen.
func fillColumn(fb []byte, w, h, x, y0, y1 int, c color.RGBA) {
	if y0 < 0 {
		y0 = 0
	}
	if y1 >= h {
		y1 = h - 1
	}
	for y := y0; y <= y1; y++ {
		setPixel(fb, w, x, y, c)
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
