package render

import (
	"image/color"
	"math"
)

// The renderer draws into a plain RGBA byte buffer (4 bytes per pixel, row
// major) which the Ebiten layer uploads in one call. Keeping the drawing here as
// pure pixel writes leaves this code free of any graphics dependency.

// Distance shading follows a simple inverse model, f = 1/(1 + dist*shadeDecay),
// clamped so far surfaces dim toward darkness but never vanish entirely.
const (
	shadeDecay  = 0.18 // how quickly brightness falls off with distance, per tile
	shadeFloor  = 0.08 // minimum brightness, so distant surfaces stay legible
	sideFaceDim = 0.72 // extra dimming on north/south faces so edges read clearly
)

// shade darkens a colour with distance (for the dim look) and a little extra for
// north/south faces so edges read clearly.
func shade(c color.RGBA, dist float64, side int) color.RGBA {
	f := 1.0 / (1.0 + dist*shadeDecay)
	if side == 1 {
		f *= sideFaceDim
	}
	f = math.Max(shadeFloor, math.Min(1, f))
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
