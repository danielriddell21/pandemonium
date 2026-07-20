package world

// HazardKind distinguishes the damaging floors a level can carry.
type HazardKind uint8

const (
	// HazardNukage is radioactive slime: a steady drain that a radiation suit
	// shrugs off entirely.
	HazardNukage HazardKind = iota
	// HazardLava is molten floor: it burns faster than slime, and no radiation
	// suit will save you from it.
	HazardLava
)

// HazardCell is the damage a tile deals and the kind of hazard it is.
type HazardCell struct {
	Rate float64
	Kind HazardKind
}

// hazard drain rates, per second.
const (
	nukageRate = 8.0
	lavaRate   = 16.0
)

// placeHazards floods a small pool of damaging floor across a patch of
// base-height tiles on some levels — radioactive slime, or, less often, molten
// lava. Hazard tiles stay walkable — they are a cost, not a barrier — so they
// never affect reachability. The pool is kept clear of the spawn so a run never
// opens standing in it.
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

	// Lava turns up on the minority of hazard levels; it is the nastier pool.
	kind, rate := HazardNukage, float64(nukageRate)
	if g.chance(0.35) {
		kind, rate = HazardLava, lavaRate
	}

	const maxPool = 14
	start := seeds[g.intn(len(seeds))]
	l.Hazard = make(map[Coord]HazardCell)
	queue := []Coord{start}
	l.Hazard[start] = HazardCell{Rate: rate, Kind: kind}
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
			l.Hazard[n] = HazardCell{Rate: rate, Kind: kind}
			queue = append(queue, n)
		}
	}
}
