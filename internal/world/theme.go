package world

const NumThemes = 3

func assignThemes(l *Level, g *rng, rooms []rect) {
	l.Theme = make([]uint8, l.Width*l.Height)
	for _, r := range rooms {
		th := uint8(g.intn(NumThemes))
		for y := r.y; y < r.y+r.h; y++ {
			for x := r.x; x < r.x+r.w; x++ {
				if l.At(x, y).Walkable() {
					l.Theme[y*l.Width+x] = th
				}
			}
		}
	}
}

func (l *Level) ThemeAt(x, y int) uint8 {
	if !l.InBounds(x, y) || len(l.Theme) == 0 {
		return 0
	}
	return l.Theme[y*l.Width+x]
}
