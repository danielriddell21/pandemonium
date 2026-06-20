package render

import (
	"image/color"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

// drawPowerupTint washes the view with a colour while a timed powerup is active:
// invulnerability reads cold and bright, the radiation suit green — mirroring
// DOOM's screen tints. Only the strongest active tint shows. Berserk lasts the
// whole level and so deliberately leaves no persistent wash.
func drawPowerupTint(fb []byte, cfg Config, g *sim.Game) {
	_ = cfg
	var c color.RGBA
	switch {
	case g.Player.Invulnerable():
		c = color.RGBA{R: 220, G: 220, B: 235}
	case g.Player.RadSuited():
		c = color.RGBA{R: 60, G: 190, B: 70}
	default:
		return
	}
	blendOver(fb, c, 0.22)
}

// blendOver alpha-blends a flat colour over the whole framebuffer.
func blendOver(fb []byte, c color.RGBA, a float64) {
	ri, gi, bi := float64(c.R)*a, float64(c.G)*a, float64(c.B)*a
	keep := 1 - a
	for i := 0; i+3 < len(fb); i += 4 {
		fb[i] = uint8(float64(fb[i])*keep + ri)
		fb[i+1] = uint8(float64(fb[i+1])*keep + gi)
		fb[i+2] = uint8(float64(fb[i+2])*keep + bi)
	}
}
