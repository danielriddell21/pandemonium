package render

import "image/color"

// MenuItem is one selectable row of a menu screen: a label and an optional value
// shown beside it (e.g. a setting and its current state).
type MenuItem struct {
	Label string
	Value string
}

// Menu renders a full-screen menu — a title, an optional subtitle, the items
// with the selected one highlighted, and a footer hint — and returns the RGBA
// buffer (owned by the Renderer, overwritten on the next call). The app drives
// selection; this only draws.
func (r *Renderer) Menu(title, subtitle string, items []MenuItem, selected int, footer string) []byte {
	fillBackground(r.fb, palette.ceiling)

	titleCol := color.RGBA{R: 222, G: 120, B: 60, A: 255}
	drawTextCentered(r.fb, r.cfg, r.cfg.Height/6, title, titleCol)
	if subtitle != "" {
		drawTextCentered(r.fb, r.cfg, r.cfg.Height/6+16, subtitle, palette.hudDiag)
	}

	// Size the value column off the widest label so values line up.
	widest := 0
	for _, it := range items {
		if len(it.Label) > widest {
			widest = len(it.Label)
		}
	}

	// Lay the items out as a block between the title and the footer, shrinking
	// the row pitch if there are enough of them to crowd the space.
	const idealRow = 18
	footerY := r.cfg.Height - 12
	top := r.cfg.Height/6 + 24
	if subtitle != "" {
		top = r.cfg.Height/6 + 44 // clear the subtitle line before the items
	}
	rowH := idealRow
	if n := len(items); n > 0 {
		if fit := (footerY - 8 - top) / n; fit < rowH {
			rowH = fit
		}
	}

	row := top
	for i, it := range items {
		line := it.Label
		if it.Value != "" {
			pad := widest - len(it.Label) + 2
			for range pad {
				line += " "
			}
			line += it.Value
		}
		col := palette.hudText
		if i == selected {
			col = titleCol
			line = "> " + line
		} else {
			line = "  " + line
		}
		drawTextCentered(r.fb, r.cfg, row, line, col)
		row += rowH
	}

	if footer != "" {
		drawTextCentered(r.fb, r.cfg, footerY, footer, palette.hudDiag)
	}
	return r.fb
}
