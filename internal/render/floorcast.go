package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

// drawFloorCeiling casts the textured floor and ceiling. For each scanline at or
// below the horizon it walks the world position from the leftmost ray to the
// rightmost ray, sampling the floor texture, and mirrors the same row above the
// horizon for the ceiling. Pairing row y (floor) with row h-1-y (ceiling) tiles
// every row of the frame, so no flat fill is needed; the walls drawn afterwards
// overwrite their columns on top. Falling back to the flat fill keeps callers
// safe if the textures are missing.
func drawFloorCeiling(fb []byte, g *sim.Game, cam camera, cfg Config, tx *textureSet) {
	if tx.floor == nil || tx.ceiling == nil {
		clearBackground(fb, cfg)
		return
	}
	w, h := cfg.Width, cfg.Height
	px, py := g.Player.Pos.X, g.Player.Pos.Y
	floor, ceil := tx.floor, tx.ceiling

	// Ray directions at the screen's left (cameraX = -1) and right (+1) edges.
	rdx0, rdy0 := cam.dirX-cam.planeX, cam.dirY-cam.planeY
	rdx1, rdy1 := cam.dirX+cam.planeX, cam.dirY+cam.planeY
	posZ := 0.5 * float64(h) // eye height in pixels: horizon at mid-screen

	for y := h / 2; y < h; y++ {
		p := float64(y - h/2)
		if p < 1 {
			p = 1 // avoid the divide-by-zero exactly on the horizon row
		}
		rowDist := posZ / p
		stepX := rowDist * (rdx1 - rdx0) / float64(w)
		stepY := rowDist * (rdy1 - rdy0) / float64(w)
		fx := px + rowDist*rdx0
		fy := py + rowDist*rdy0

		ff := shadeFactor(rowDist) // distance dimming, matching the walls
		cf := ff * 0.85            // ceiling a touch darker than the floor
		cy := h - 1 - y            // mirrored ceiling row

		for x := 0; x < w; x++ {
			tcx := int(float64(floor.w) * (fx - math.Floor(fx)))
			tcy := int(float64(floor.h) * (fy - math.Floor(fy)))
			setPixel(fb, w, x, y, scaleColor(floor.at(tcx, tcy), ff))
			setPixel(fb, w, x, cy, scaleColor(ceil.at(tcx, tcy), cf))
			fx += stepX
			fy += stepY
		}
	}
}

// shadeFactor is the distance-dimming multiplier used for flat surfaces, matching
// the wall shading curve (see shade) for a north/south face.
func shadeFactor(dist float64) float64 {
	return math.Max(0.08, math.Min(1, 1.0/(1.0+dist*0.18)))
}

// scaleColor multiplies an RGB colour by f, keeping it opaque.
func scaleColor(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: 255,
	}
}
