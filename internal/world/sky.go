package world

const skyHeadroom = 3.0

func assignSky(l *Level, g *rng, rooms []rect) {
	for _, r := range rooms {
		if g.intn(10) >= 3 { // roughly three rooms in ten open to the sky
			continue
		}
		for y := r.y; y < r.y+r.h; y++ {
			for x := r.x; x < r.x+r.w; x++ {
				if !l.At(x, y).Walkable() {
					continue
				}
				i := y*l.Width + x
				l.Sky[i] = true
				l.Light[i] = 1
				l.setCeil(x, y, l.Floor(x, y)+skyHeadroom)
			}
		}
	}
}
