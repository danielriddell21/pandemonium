package render

// drawCrosshair paints a small four-armed cross at the centre of the world view
// (which sits above the status bar unless the HUD is hidden).
func drawCrosshair(fb []byte, cfg Config, bareView bool) {
	viewH := cfg.Height
	if !bareView {
		viewH -= StatusBarH
	}
	cx, cy := cfg.Width/2, viewH/2
	const gap, arm = 2, 3
	c := palette.hudText
	for i := gap; i < gap+arm; i++ {
		setPixel(fb, cfg.Width, cx+i, cy, c)
		setPixel(fb, cfg.Width, cx-i, cy, c)
		setPixel(fb, cfg.Width, cx, cy+i, c)
		setPixel(fb, cfg.Width, cx, cy-i, c)
	}
}
