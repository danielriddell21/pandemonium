package render

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/danielriddell21/pandemonium/internal/hud"
)

const (
	hudMarginX  = 4  // left inset of the message line, in pixels
	hudBaseline = 13 // text baseline from the top, in pixels
)

// drawMessage renders a single message into the framebuffer at the top left, with
// a one-pixel drop shadow so it stays legible over any background. The colour
// depends on the channel.
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

// glyphWidth is the fixed advance of basicfont.Face7x13, used to centre text.
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

// drawTextCentered draws s horizontally centred on the framebuffer at baseline y,
// with a one-pixel drop shadow for legibility.
func drawTextCentered(fb []byte, cfg Config, y int, s string, c color.RGBA) {
	dst := framebufferImage(fb, cfg)
	x := (cfg.Width - len(s)*glyphWidth) / 2
	drawText(dst, x+1, y+1, s, palette.hudDrop)
	drawText(dst, x, y, s, c)
}

// drawNotice shows a transient gameplay message (pickups, keys, secrets) centred
// low on the screen, above the weapon.
func drawNotice(fb []byte, cfg Config, msg string) {
	if msg == "" {
		return
	}
	drawTextCentered(fb, cfg, cfg.Height*3/4, msg, palette.hudText)
}

// framebufferImage wraps a framebuffer slice as an image.RGBA for font drawing.
func framebufferImage(fb []byte, cfg Config) *image.RGBA {
	return &image.RGBA{
		Pix:    fb,
		Stride: cfg.Width * 4,
		Rect:   image.Rect(0, 0, cfg.Width, cfg.Height),
	}
}
