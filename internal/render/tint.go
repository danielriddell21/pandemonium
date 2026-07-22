package render

import (
	"image/color"

	"github.com/danielriddell21/crucible/paint"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

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
	paint.BlendOver(fb, c, 0.22)
}
