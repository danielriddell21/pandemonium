package world

import "github.com/danielriddell21/crucible/worldgen"

type HazardKind uint8

const (
	HazardNukage HazardKind = iota

	HazardLava
)

type HazardCell struct {
	Rate float64
	Kind HazardKind
}

const (
	nukageRate = 8.0
	lavaRate   = 16.0
)

func placeHazards(l *Level, g *worldgen.RNG) {
	if !g.Chance(0.5) {
		return // only some levels carry a pool
	}
	seeds := hazardSeeds(l)
	if len(seeds) == 0 {
		return
	}

	// Lava turns up on the minority of hazard levels; it is the nastier pool.
	kind, rate := HazardNukage, float64(nukageRate)
	if g.Chance(0.35) {
		kind, rate = HazardLava, lavaRate
	}
	// Start from a random candidate and take the first that can hold a pool
	// without cutting the level off.
	first := g.IntN(len(seeds))
	for i := range seeds {
		if floodHazard(l, seeds[(first+i)%len(seeds)], kind, rate) {
			return
		}
	}
}

func hazardSeeds(l *Level) []Coord {
	var seeds []Coord
	for y := range l.H {
		for x := range l.W {
			c := Coord{X: x, Y: y}
			if l.At(x, y) == TileFloor && l.Floor(x, y) == 0 && c != l.Exit && cheby(c, l.Spawn) >= 4 {
				seeds = append(seeds, c)
			}
		}
	}
	return seeds
}

func floodHazard(l *Level, start Coord, kind HazardKind, rate float64) bool {
	const maxPool = 14
	cell := HazardCell{Rate: rate, Kind: kind}
	l.Hazard = make(map[Coord]HazardCell)
	if !wet(l, start, cell) {
		return false
	}
	queue := []Coord{start}
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
			if wet(l, n, cell) {
				queue = append(queue, n)
			}
		}
	}
	return true
}

func wet(l *Level, c Coord, cell HazardCell) bool {
	// Wading is always a choice: a cell joins the pool only if a dry route to
	// the exit and to every keycard survives it.
	l.Hazard[c] = cell
	if dryRoutes(l) {
		return true
	}
	delete(l.Hazard, c)
	return false
}

func dryRoutes(l *Level) bool {
	dry := func(solid solidFn) solidFn {
		return func(c Coord) bool {
			_, pool := l.Hazard[c]
			return pool || solid(c)
		}
	}
	if !reachable(l, l.Spawn, l.Exit, dry(blocksWalls(l))) {
		return false
	}
	for lock, key := range l.Locks {
		at, ok := keyCell(l, key)
		if !ok || !reachable(l, l.Spawn, at, dry(lockedSolid(l, lock))) {
			return false
		}
	}
	return true
}
