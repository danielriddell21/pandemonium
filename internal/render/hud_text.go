package render

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/danielriddell21/crucible/hud"
	"github.com/danielriddell21/crucible/keymap"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

const (
	hudMarginX  = 4
	hudBaseline = 13
)

func drawMessage(fb []byte, cfg Config, msg string, ch hud.Channel) {
	if msg == "" {
		return
	}
	dst := &image.RGBA{
		Pix:    fb,
		Stride: cfg.Width * 4,
		Rect:   image.Rect(0, 0, cfg.Width, cfg.Height),
	}
	fg := palette.hudText
	if ch == hud.Diagnostic {
		fg = palette.hudDiag
	}
	drawText(dst, hudMarginX+1, hudBaseline+1, msg, palette.hudDrop)
	drawText(dst, hudMarginX, hudBaseline, msg, fg)
}

const glyphWidth = 7

func drawText(dst *image.RGBA, x, y int, s string, c color.RGBA) {
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(s)
}

func drawTextCentered(fb []byte, cfg Config, y int, s string, c color.RGBA) {
	dst := framebufferImage(fb, cfg)
	x := (cfg.Width - len(s)*glyphWidth) / 2
	drawText(dst, x+1, y+1, s, palette.hudDrop)
	drawText(dst, x, y, s, c)
}

const (
	plateDim  = 0.55
	platePadX = 6
)

func drawPlate(fb []byte, cfg Config, x, y, textW int) {
	// A dimmed strip behind centre-screen text keeps it legible over the
	// weapon and bright walls without hiding either.
	for py := y - hudBaseline + 2; py <= y+4; py++ {
		for px := x - platePadX; px < x+textW+platePadX; px++ {
			if px >= 0 && px < cfg.Width && py >= 0 && py < cfg.Height {
				blendPixel(fb, cfg.Width, px, py, palette.hudDrop, plateDim)
			}
		}
	}
}

func drawNotice(fb []byte, cfg Config, msg string) {
	if msg == "" {
		return
	}
	y := cfg.Height * 3 / 4
	drawPlate(fb, cfg, (cfg.Width-len(msg)*glyphWidth)/2, y, len(msg)*glyphWidth)
	drawTextCentered(fb, cfg, y, msg, palette.hudText)
}

func drawInteractPrompt(fb []byte, cfg Config, g *sim.Game) {
	var action string
	switch g.FacingInteractable() {
	case sim.InteractDoor:
		action = "OPEN DOOR"
	case sim.InteractSwitch:
		action = "PRESS SWITCH"
	case sim.InteractPanel:
		action = "SILENCE ALARM"
		if g.PanelSpent() {
			return
		}
	default:
		return
	}
	face := keymap.Face{LineHeight: hudBaseline, Measure: func(s string) int { return len(s) * glyphWidth }}
	line := keymap.CenterPrompt(keymap.Binding{Key: "E", Action: action}, cfg.Width, cfg.Height, face)
	dst := framebufferImage(fb, cfg)
	drawPlate(fb, cfg, line.X, line.Y, len(line.Text)*glyphWidth)
	drawText(dst, line.X+1, line.Y+1, line.Text, palette.hudDrop)
	drawText(dst, line.X, line.Y, line.Text, palette.hudText)
}

func framebufferImage(fb []byte, cfg Config) *image.RGBA {
	return &image.RGBA{
		Pix:    fb,
		Stride: cfg.Width * 4,
		Rect:   image.Rect(0, 0, cfg.Width, cfg.Height),
	}
}
