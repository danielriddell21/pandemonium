package render

import (
	"fmt"
	"image/color"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// statusBarH is the height in pixels of the bottom status panel.
const statusBarH = 38

// statusbar colours.
var (
	barBG     = color.RGBA{R: 30, G: 26, B: 24, A: 255}
	barBorder = color.RGBA{R: 84, G: 70, B: 56, A: 255}
	barLabel  = color.RGBA{R: 150, G: 140, B: 120, A: 255}
	armorText = color.RGBA{R: 150, G: 190, B: 220, A: 255}
)

// keyColors mirror the keycard sprites, indexed for red/blue/yellow.
var keyColors = [...]struct {
	kind world.ItemKind
	on   color.RGBA
}{
	{world.ItemKeyRed, color.RGBA{R: 210, G: 50, B: 50, A: 255}},
	{world.ItemKeyBlue, color.RGBA{R: 70, G: 110, B: 220, A: 255}},
	{world.ItemKeyYellow, color.RGBA{R: 220, G: 200, B: 60, A: 255}},
}

// drawStatusBar paints the bottom HUD panel: the player's face, health, armour,
// current-weapon ammo and the keycards held. It replaces the minimal health bar.
func drawStatusBar(fb []byte, cfg Config, g *sim.Game, tx *textureSet) {
	w, h := cfg.Width, cfg.Height
	top := h - statusBarH

	// Panel and top edge.
	fillRows(fb, w, top, h, barBG)
	for x := range w {
		setPixel(fb, w, x, top, barBorder)
	}

	dst := framebufferImage(fb, cfg)
	labelY := top + 14
	valueY := top + 29

	// Health (left), coloured by how much remains.
	hp := clampFrac(g.Player.Health / sim.MaxHealth)
	drawText(dst, 12, labelY, "HEALTH", barLabel)
	drawText(dst, 12, valueY, fmt.Sprintf("%d%%", int(hp*100+0.5)), healthColor(hp))

	// Armour.
	ar := clampFrac(g.Player.Armor / sim.MaxArmor)
	drawText(dst, 92, labelY, "ARMOR", barLabel)
	drawText(dst, 92, valueY, fmt.Sprintf("%d%%", int(ar*100+0.5)), armorText)

	// Face, centred.
	if len(tx.face) > 0 {
		face := tx.face[faceBand(hp)]
		fs := statusBarH - 6
		blitTexture(fb, cfg, face, w/2-fs/2, top+3, fs, fs)
	}

	// Ammo for the current weapon (right of centre).
	ax := w/2 + 60
	drawText(dst, ax, labelY, "AMMO", barLabel)
	drawText(dst, ax, valueY, ammoText(g.Player), palette.hudText)

	// Keycards held (far right): lit pips for those collected, dim outlines else.
	kx := w - 12 - len(keyColors)*14
	for i, k := range keyColors {
		x0 := kx + i*14
		drawKeyPip(fb, cfg, x0, top+11, k.on, g.Player.HasKey(k.kind))
	}
}

// ammoText renders the current weapon's ammo count, or a dash for the fists.
func ammoText(p sim.Player) string {
	switch p.Weapon {
	case sim.Pistol:
		return fmt.Sprintf("%d", p.Bullets)
	case sim.Shotgun:
		return fmt.Sprintf("%d", p.Shells)
	default:
		return "--"
	}
}

// drawKeyPip draws a small keycard indicator: filled if held, a dim outline if not.
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
	dim := scaleColor(on, 0.28)
	for x := x0; x < x0+pw; x++ {
		setPixel(fb, cfg.Width, x, y0, dim)
		setPixel(fb, cfg.Width, x, y0+ph-1, dim)
	}
	for y := y0; y < y0+ph; y++ {
		setPixel(fb, cfg.Width, x0, y, dim)
		setPixel(fb, cfg.Width, x0+pw-1, y, dim)
	}
}

// faceBand maps a health fraction to a face index (0 healthy .. 3 dead).
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
