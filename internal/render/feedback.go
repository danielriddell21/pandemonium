package render

import (
	"image/color"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

var (
	painRed   = color.RGBA{R: 170, G: 18, B: 14, A: 255}
	hitWhite  = color.RGBA{R: 255, G: 246, B: 224, A: 255}
	alertCalm = color.RGBA{R: 196, G: 170, B: 60, A: 255}
	alertHot  = color.RGBA{R: 232, G: 64, B: 40, A: 255}
	alertHush = color.RGBA{R: 110, G: 170, B: 210, A: 255}
	barTrough = color.RGBA{R: 22, G: 18, B: 18, A: 255}
)

const (
	painReach   = 0.75
	painSide    = 0.5
	hotAlert    = 0.75
	threatBarH  = 6
	threatTop   = 20
	threatLabel = 15
)

func drawFeedback(fb []byte, cfg Config, g *sim.Game) {
	viewH := cfg.Height - StatusBarH
	if pain := g.Pain(); pain > 0 {
		drawPain(fb, cfg, viewH, pain, g.Player.FaceDir())
	}
	if flash := g.AlarmFlash(); flash > 0 && g.Tick64()/6%2 == 0 {
		drawEdgeBars(fb, cfg, viewH, alertHot, flash)
	}
	if g.HitMarker() {
		drawHitMarker(fb, cfg, viewH)
	}
}

func drawPain(fb []byte, cfg Config, viewH int, pain float64, side int) {
	w := cfg.Width
	for y := range viewH {
		ny := float64(2*y)/float64(viewH) - 1
		for x := range w {
			nx := float64(2*x)/float64(w) - 1
			// Red closes in from the edges, never over the middle of the view,
			// and leans to the side the hit came from.
			edge := max(nx*nx, ny*ny)
			if edge < 1-painReach {
				continue
			}
			a := pain * painReach * (edge - (1 - painReach)) / painReach
			if side != 0 && nx*float64(side) < 0 {
				a *= painSide
			}
			blendPixel(fb, w, x, y, painRed, a)
		}
	}
}

func drawEdgeBars(fb []byte, cfg Config, viewH int, c color.RGBA, strength float64) {
	thick := max(cfg.Height/60, 2)
	for y := range thick {
		for x := range cfg.Width {
			blendPixel(fb, cfg.Width, x, y, c, strength)
			blendPixel(fb, cfg.Width, x, viewH-1-y, c, strength)
		}
	}
}

func drawHitMarker(fb []byte, cfg Config, viewH int) {
	cx, cy := cfg.Width/2, viewH/2
	const gap, arm = 3, 4
	for i := gap; i < gap+arm; i++ {
		for _, d := range [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
			setPixel(fb, cfg.Width, cx+i*d[0], cy+i*d[1], hitWhite)
		}
	}
}

func drawThreatBar(fb []byte, cfg Config, g *sim.Game) {
	if st, ok := g.Boss(); ok {
		drawBossBar(fb, cfg, st, g.Tick64())
		return
	}
	level := g.Alert()
	switch {
	case g.AlarmQuiet():
		drawMeter(fb, cfg, "ALARM SILENCED", 0, alertHush, alertHush)
	case level >= hotAlert:
		label := "ALERT"
		if g.Tick64()/8%2 == 0 {
			label = "ALERT!"
		}
		drawMeter(fb, cfg, label, level, alertHot, alertHot)
	case level > 0.02 || g.Alarms() > 0:
		drawMeter(fb, cfg, "ALERT", level, alertCalm, palette.hudText)
	}
}

func drawBossBar(fb []byte, cfg Config, st sim.BossStatus, tick uint64) {
	fill := [...]color.RGBA{{R: 214, G: 120, B: 40, A: 255}, {R: 222, G: 70, B: 40, A: 255}, {R: 214, G: 40, B: 96, A: 255}}[st.Phase]
	label, text := "OVERLORD", palette.hudText
	switch st.Action {
	case sim.BossVolley:
		label, text = "VOLLEY", tintWindup
	case sim.BossSummon:
		label, text = "SUMMONING", tintWindup
	case sim.BossCharge, sim.BossCharging:
		label, text = "CHARGE!", tintCharge
	case sim.BossStunned:
		label, text = "STUNNED", tintStun
		if tick/6%2 == 0 {
			fill = tintStun // open to double damage: make the window impossible to miss
		}
	}
	x0, bw := drawMeter(fb, cfg, label, st.Health/st.MaxHealth, fill, text)
	// Notches where the next phase begins.
	for _, frac := range [2]float64{1.0 / 3, 2.0 / 3} {
		x := x0 + int(frac*float64(bw))
		for y := threatTop; y < threatTop+threatBarH; y++ {
			setPixel(fb, cfg.Width, x, y, barTrough)
		}
	}
}

func drawMeter(fb []byte, cfg Config, label string, frac float64, fill, text color.RGBA) (x0, bw int) {
	bw = min(cfg.Width/2, 220)
	x0 = (cfg.Width - bw) / 2
	drawTextCentered(fb, cfg, threatLabel, label, text)
	fillBox(fb, cfg, x0-1, threatTop-1, bw+2, threatBarH+2, barBorder)
	fillBox(fb, cfg, x0, threatTop, bw, threatBarH, barTrough)
	fillBox(fb, cfg, x0, threatTop, int(clampFrac(frac)*float64(bw)), threatBarH, fill)
	return x0, bw
}

func blendPixel(fb []byte, w, x, y int, c color.RGBA, a float64) {
	i := (y*w + x) * 4
	if a <= 0 || i < 0 || i+3 >= len(fb) {
		return
	}
	a = min(a, 1)
	fb[i] = uint8(float64(fb[i]) + (float64(c.R)-float64(fb[i]))*a)
	fb[i+1] = uint8(float64(fb[i+1]) + (float64(c.G)-float64(fb[i+1]))*a)
	fb[i+2] = uint8(float64(fb[i+2]) + (float64(c.B)-float64(fb[i+2]))*a)
	fb[i+3] = 255
}
