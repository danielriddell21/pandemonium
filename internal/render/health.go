package render

import "image/color"

const (
	healthBarW      = 80
	healthBarH      = 6
	healthBarMargin = 6
)

// drawHealthBar draws a small health gauge in the bottom-left corner. frac is the
// player's health fraction in [0,1]; the fill shifts from red (low) to green.
func drawHealthBar(fb []byte, cfg Config, frac float64) {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	x0 := healthBarMargin
	y0 := cfg.Height - healthBarMargin - healthBarH

	for y := y0; y < y0+healthBarH; y++ {
		for x := x0; x < x0+healthBarW; x++ {
			setPixel(fb, cfg.Width, x, y, palette.hudDrop)
		}
	}

	fill := int(float64(healthBarW) * frac)
	c := healthColor(frac)
	for y := y0; y < y0+healthBarH; y++ {
		for x := x0; x < x0+fill; x++ {
			setPixel(fb, cfg.Width, x, y, c)
		}
	}
}

// healthColor blends red (empty) toward green (full).
func healthColor(frac float64) color.RGBA {
	return color.RGBA{
		R: uint8(40 + 200*(1-frac)),
		G: uint8(40 + 180*frac),
		B: 50,
		A: 255,
	}
}
