package world

import "github.com/danielriddell21/crucible/worldgen"

func assignLight(l *Level, g *worldgen.RNG, rooms []rect) {
	for _, r := range rooms {
		bright := roomBrightness(g)
		if r.Contains(l.Spawn) {
			bright = 1.0
		}
		for y := r.Y; y < r.Y+r.H; y++ {
			for x := r.X; x < r.X+r.W; x++ {
				if l.At(x, y).Walkable() {
					l.Light[y*l.W+x] = bright
				}
			}
		}
	}
	for c := range l.Hazard {
		if l.InBounds(c.X, c.Y) {
			l.Light[c.Y*l.W+c.X] = 0.85 // slime is faintly self-lit
		}
	}
}

func roomBrightness(g *worldgen.RNG) float64 {
	switch g.IntN(5) {
	case 0:
		return 0.55 // dim
	case 1:
		return 0.75 // shadowed
	default:
		return 1.0 // lit
	}
}
