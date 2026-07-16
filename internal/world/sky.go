package world

// skyHeadroom is how tall an open-air room's ceiling stands above its floor, so
// a sky room reads as an open courtyard rather than a low glowing roof.
const skyHeadroom = 3.0

// assignSky opens some rooms to the sky: their ceilings lift to an open height,
// their light goes full, and their cells are flagged so the renderer paints open
// air instead of stone overhead. It runs last and only touches ceiling height,
// light and the sky flag, so it never affects reachability or any earlier,
// rng-driven placement.
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
