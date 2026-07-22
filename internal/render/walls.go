package render

import (
	"image/color"

	"github.com/danielriddell21/crucible/paint"
)

const (
	shadeDecay  = 0.18
	shadeFloor  = 0.08
	sideFaceDim = 0.72
)

func shade(c color.RGBA, dist float64, side int) color.RGBA {
	f := 1.0 / (1.0 + dist*shadeDecay)
	if side == 1 {
		f *= sideFaceDim
	}
	f = max(shadeFloor, min(1, f))
	return paint.Scale(c, f)
}

func fillRows(fb []byte, w, y0, y1 int, c color.RGBA) {
	for y := y0; y < y1; y++ {
		for x := range w {
			setPixel(fb, w, x, y, c)
		}
	}
}

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
