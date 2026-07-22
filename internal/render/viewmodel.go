package render

import "math"

func drawViewmodel(fb []byte, cfg Config, weapon, flash *texture, firing bool, tick float64, bottom int) {
	if weapon == nil {
		return
	}
	vw := cfg.Width * 2 / 5
	vh := vw
	bob := int(3 * math.Sin(tick*0.15))
	x0 := cfg.Width/2 - vw/2
	y0 := bottom - vh + bob

	// The weapon is drawn first; the flash sits over its muzzle (centred
	// horizontally, near the top of the sprite where the barrel points).
	blitTexture(fb, cfg, weapon, x0, y0, vw, vh)
	if firing && flash != nil {
		fw := vw / 2
		fy := y0 + int(float64(vh)*0.18)
		blitTexture(fb, cfg, flash, cfg.Width/2-fw/2, fy, fw, fw)
	}
}

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
