package world

// NumThemes is how many visual wall themes a level draws from; the renderer keeps
// a matching set of tinted textures.
const NumThemes = 3

// assignThemes gives each room one of a few wall themes so the level reads as
// distinct areas rather than one endless texture. Corridors keep the base theme.
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

// ThemeAt returns the wall theme at (x, y); out of bounds is the base theme.
func (l *Level) ThemeAt(x, y int) uint8 {
	if !l.InBounds(x, y) || len(l.Theme) == 0 {
		return 0
	}
	return l.Theme[y*l.Width+x]
}
