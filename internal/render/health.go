package render

import "image/color"

// healthColor blends red (empty) toward green (full).
func healthColor(frac float64) color.RGBA {
	return color.RGBA{
		R: uint8(40 + 200*(1-frac)),
		G: uint8(40 + 180*frac),
		B: 50,
		A: 255,
	}
}
