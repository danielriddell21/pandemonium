package world

// assignLight gives each room its own brightness so the level reads as a series
// of differently-lit spaces rather than one evenly-shaded maze. Corridors take a
// middling light, and hazard pools glow a little so slime stays visible. The
// spawn room stays bright so a run never opens in gloom.
func assignLight(l *Level, g *rng, rooms []rect) {
	for _, r := range rooms {
		level := roomBrightness(g)
		if r.contains(l.Spawn) {
			level = 1.0
		}
		for y := r.y; y < r.y+r.h; y++ {
			for x := r.x; x < r.x+r.w; x++ {
				if l.At(x, y).Walkable() {
					l.Light[y*l.Width+x] = level
				}
			}
		}
	}
	for c := range l.Hazard {
		if l.InBounds(c.X, c.Y) {
			l.Light[c.Y*l.Width+c.X] = 0.85 // slime is faintly self-lit
		}
	}
}

// roomBrightness picks a per-room light level, biased toward well-lit rooms with
// the occasional gloomy one.
func roomBrightness(g *rng) float64 {
	switch g.intn(5) {
	case 0:
		return 0.55 // dim
	case 1:
		return 0.75 // shadowed
	default:
		return 1.0 // lit
	}
}
