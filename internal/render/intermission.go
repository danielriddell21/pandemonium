package render

import (
	"fmt"
	"image/color"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

func drawIntermission(fb []byte, cfg Config, s sim.LevelStats) {
	fillBackground(fb, palette.ceiling)

	title := color.RGBA{R: 222, G: 120, B: 60, A: 255}
	value := palette.hudText

	row := cfg.Height/2 - 48
	drawTextCentered(fb, cfg, row, "LEVEL COMPLETE", title)

	row += 28
	drawTextCentered(fb, cfg, row, fmt.Sprintf("KILLS    %3d%%", s.KillsPct()), value)
	row += 18
	drawTextCentered(fb, cfg, row, fmt.Sprintf("ITEMS    %3d%%", s.ItemsPct()), value)
	row += 18
	drawTextCentered(fb, cfg, row, fmt.Sprintf("SECRETS  %3d%%", s.SecretsPct()), value)
	row += 18
	timeColor := value
	if s.Par > 0 && s.Elapsed <= s.Par { // beat par — call it out in green
		timeColor = palette.hudDiag
	}
	drawTextCentered(fb, cfg, row, "TIME     "+formatTime(s.Elapsed), timeColor)
	row += 18
	drawTextCentered(fb, cfg, row, "PAR      "+formatTime(s.Par), value)

	drawTextCentered(fb, cfg, cfg.Height-24, "Press Enter to continue", palette.hudDiag)
}

func formatTime(secs float64) string {
	total := int(secs)
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

func fillBackground(fb []byte, c color.RGBA) {
	for i := 0; i < len(fb); i += 4 {
		fb[i], fb[i+1], fb[i+2], fb[i+3] = c.R, c.G, c.B, 255
	}
}
