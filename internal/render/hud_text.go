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

func drawNotice(fb []byte, cfg Config, msg string) {
	if msg == "" {
		return
	}
	drawTextCentered(fb, cfg, cfg.Height*3/4, msg, palette.hudText)
}

func framebufferImage(fb []byte, cfg Config) *image.RGBA {
	return &image.RGBA{
		Pix:    fb,
		Stride: cfg.Width * 4,
		Rect:   image.Rect(0, 0, cfg.Width, cfg.Height),
	}
}
