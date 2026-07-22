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
	floodHazard(l, seeds[g.IntN(len(seeds))], kind, rate)
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

func floodHazard(l *Level, start Coord, kind HazardKind, rate float64) {
	const maxPool = 14
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
