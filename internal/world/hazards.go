package world

// hazardRate is how much health a nukage tile drains per second, in the spirit of
// DOOM's slime floors.
const hazardRate = 8.0

// placeHazards floods a small pool of damaging "nukage" floor across a patch of
// base-height tiles on some levels. Hazard tiles stay walkable — they are a cost,
// not a barrier — so they never affect reachability. The pool is kept clear of
// the spawn so a run never opens standing in slime.
func placeHazards(l *Level, g *rng) {
	if !g.chance(0.5) {
		return // only some levels carry a pool
	}
	var seeds []Coord
	for y := range l.Height {
		for x := range l.Width {
			c := Coord{X: x, Y: y}
			if l.At(x, y) == TileFloor && l.Floor(x, y) == 0 && c != l.Exit && cheby(c, l.Spawn) >= 4 {
				seeds = append(seeds, c)
			}
		}
	}
	if len(seeds) == 0 {
		return
	}

	const maxPool = 14
	start := seeds[g.intn(len(seeds))]
	l.Hazard = make(map[Coord]float64)
	queue := []Coord{start}
	l.Hazard[start] = hazardRate
	for len(queue) > 0 && len(l.Hazard) < maxPool {
		c := queue[0]
		queue = queue[1:]
		for _, n := range neighbors4(c) {
			// Spread only across same-height open floor, so the pool sits flat.
			if l.At(n.X, n.Y) != TileFloor || l.Floor(n.X, n.Y) != 0 {
				continue
			}
			if _, ok := l.Hazard[n]; ok || cheby(n, l.Spawn) < 4 {
				continue
			}
			l.Hazard[n] = hazardRate
			queue = append(queue, n)
		}
	}
}
