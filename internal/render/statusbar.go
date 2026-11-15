package render

import (
	"fmt"
	"image/color"

	"github.com/danielriddell21/crucible/paint"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

const StatusBarH = 38

var (
	barBG     = color.RGBA{R: 30, G: 26, B: 24, A: 255}
	barBorder = color.RGBA{R: 84, G: 70, B: 56, A: 255}
	barLabel  = color.RGBA{R: 150, G: 140, B: 120, A: 255}
	armorText = color.RGBA{R: 150, G: 190, B: 220, A: 255}
)

var keyColors = [...]struct {
	kind world.ItemKind
	on   color.RGBA
}{
	{world.ItemKeyRed, color.RGBA{R: 210, G: 50, B: 50, A: 255}},
	{world.ItemKeyBlue, color.RGBA{R: 70, G: 110, B: 220, A: 255}},
	{world.ItemKeyYellow, color.RGBA{R: 220, G: 200, B: 60, A: 255}},
}

func drawStatusBar(fb []byte, cfg Config, g *sim.Game, tx *textureSet) {
	w, h := cfg.Width, cfg.Height
	top := h - StatusBarH

	// Panel and top edge.
	fillRows(fb, w, top, h, barBG)
	for x := range w {
		setPixel(fb, w, x, top, barBorder)
	}

	dst := framebufferImage(fb, cfg)
	labelY := top + 14
	valueY := top + 29

	// Columns are placed by share of the width, so the bar keeps its spacing
	// at any render size instead of letting the face and keycards sit on text.
	fs := StatusBarH - 6
	kx := w - w/32 - len(keyColors)*keyPitch
	cols := barColumns(w, fs, kx)

	// Health (left), coloured by how much remains; overheal reads past 100.
	hp := clampFrac(g.Player.Health / sim.MaxHealth)
	drawText(dst, cols.health, labelY, "HEALTH", barLabel)
	drawText(dst, cols.health, valueY, fmt.Sprintf("%d%%", int(g.Player.Health+0.5)), healthColor(hp))

	// Armour.
	drawText(dst, cols.armor, labelY, "ARMOR", barLabel)
	drawText(dst, cols.armor, valueY, fmt.Sprintf("%d%%", int(g.Player.Armor+0.5)), armorText)

	// Face, centred, gazing toward recent damage.
	if idx := faceIndex(faceBand(hp), g.Player.FaceDir()); idx >= 0 && idx < len(tx.face) {
		blitTexture(fb, cfg, tx.face[idx], w/2-fs/2, top+3, fs, fs)
	}

	// Ammo for the current weapon (right of centre).
	drawText(dst, cols.ammo, labelY, "AMMO", barLabel)
	drawText(dst, cols.ammo, valueY, ammoText(g.Player), palette.hudText)

	// Keycards held (far right): lit pips for those collected, dim outlines else.
	for i, k := range keyColors {
		drawKeyPip(fb, cfg, kx+i*keyPitch, top+11, k.on, g.Player.HasKey(k.kind))
	}
	for _, x := range [...]int{cols.armor - barGutter, w/2 - fs/2 - barGutter, w/2 + fs/2 + barGutter, kx - barGutter} {
		for y := top + 5; y < h-4; y++ {
			setPixel(fb, w, x, y, barBorder)
		}
	}
}

const (
	keyPitch  = 14
	barGutter = 5
)

type barLayout struct {
	health, armor, ammo int
}

func barColumns(w, face, keys int) barLayout {
	left := w / 26
	// Armour sits midway between the health column and the face; ammo midway
	// between the face and the keycards, never closer than a gutter to either.
	faceL, faceR := w/2-face/2, w/2+face/2
	armor := left + (faceL-left)/2
	ammo := faceR + 2*barGutter + max((keys-faceR-4*barGutter-4*glyphWidth)/2, 0)
	return barLayout{health: left, armor: armor, ammo: ammo}
}

func ammoText(p sim.Player) string {
	switch p.Weapon {
	case sim.Pistol, sim.Chaingun:
		return fmt.Sprintf("%d", p.Bullets)
	case sim.Shotgun:
		return fmt.Sprintf("%d", p.Shells)
	case sim.RocketLauncher:
		return fmt.Sprintf("%d", p.Rockets)
	default:
		return "--"
	}
}

func drawKeyPip(fb []byte, cfg Config, x0, y0 int, on color.RGBA, held bool) {
	const pw, ph = 9, 14
	if held {
		for y := y0; y < y0+ph; y++ {
			for x := x0; x < x0+pw; x++ {
				setPixel(fb, cfg.Width, x, y, on)
			}
		}
		return
	}
	dim := paint.Scale(on, 0.28)
	for x := x0; x < x0+pw; x++ {
		setPixel(fb, cfg.Width, x, y0, dim)
		setPixel(fb, cfg.Width, x, y0+ph-1, dim)
	}
	for y := y0; y < y0+ph; y++ {
		setPixel(fb, cfg.Width, x0, y, dim)
		setPixel(fb, cfg.Width, x0+pw-1, y, dim)
	}
}

func faceBand(frac float64) int {
	switch {
	case frac > 0.66:
		return 0
	case frac > 0.33:
		return 1
	case frac > 0:
		return 2
	default:
		return 3
	}
}

func clampFrac(f float64) float64 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}
