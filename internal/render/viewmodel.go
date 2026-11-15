package render

import "math"

const (
	bobRate = 3.2
	bobSway = 0.012
	bobLift = 0.008
	recoil  = 0.02
)

func drawViewmodel(fb []byte, cfg Config, weapon, flash *texture, firing bool, stride float64, bottom int) {
	if weapon == nil {
		return
	}
	vw := cfg.Width * 2 / 5
	vh := vw
	// The weapon swings with the player's stride and rests when they stand
	// still, then kicks back down as it fires.
	phase := stride * bobRate
	sway := int(float64(cfg.Width) * bobSway * math.Sin(phase))
	lift := int(float64(cfg.Width) * bobLift * math.Abs(math.Cos(phase)))
	x0 := cfg.Width/2 - vw/2 + sway
	y0 := bottom - vh + lift
	if firing {
		y0 += int(float64(cfg.Width) * recoil)
	}

	// The weapon is drawn first; the flash sits over its muzzle (centred
	// horizontally, near the top of the sprite where the barrel points).
	blitTexture(fb, cfg, weapon, x0, y0, vw, vh)
	if firing && flash != nil {
		fw := vw * 2 / 5
		fy := y0 + int(float64(vh)*0.12)
		blitTexture(fb, cfg, flash, cfg.Width/2-fw/2+sway, fy, fw, fw)
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
