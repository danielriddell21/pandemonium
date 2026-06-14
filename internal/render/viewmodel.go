package render

import "math"

// drawViewmodel blits the current weapon at the bottom-centre of the screen with
// a gentle bob, plus a muzzle flash above it while firing. tick drives the bob.
func drawViewmodel(fb []byte, cfg Config, weapon, flash *texture, firing bool, tick float64) {
	if weapon == nil {
		return
	}
	vw := cfg.Width * 2 / 5
	vh := vw
	bob := int(3 * math.Sin(tick*0.15))
	x0 := cfg.Width/2 - vw/2
	y0 := cfg.Height - vh + bob

	if firing && flash != nil {
		fw := vw / 2
		blitTexture(fb, cfg, flash, cfg.Width/2-fw/2, y0-fw/3, fw, fw)
	}
	blitTexture(fb, cfg, weapon, x0, y0, vw, vh)
}

// blitTexture nearest-samples tex into the screen rect (dx,dy,dw,dh), skipping
// transparent texels.
func blitTexture(fb []byte, cfg Config, tex *texture, dx, dy, dw, dh int) {
	for y := range dh {
		sy := y * tex.h / dh
		for x := range dw {
			texel := tex.at(x*tex.w/dw, sy)
			if texel.A < 128 {
				continue
			}
			setPixel(fb, cfg.Width, dx+x, dy+y, texel)
		}
	}
}
