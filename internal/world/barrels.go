package world

import "github.com/danielriddell21/crucible/worldgen"

func placeBarrels(l *Level, g *worldgen.RNG) {
	floors := freeFloors(l) // already excludes spawn area, exit and item cells

	// Barrels are solid, so keep them out of single-file corridors where one
	// would wall off a mandatory path; place them only in open cells with room
	// to walk around.
	open := floors[:0]
	for _, c := range floors {
		if len(walkableNeighbors(l, c)) >= 3 {
			open = append(open, c)
		}
	}
	shuffleCoords(g, open)

	n := len(open) / 30
	if n < 1 {
		n = 1
	}
	if n > 6 {
		n = 6
	}
	if n > len(open) {
		n = len(open)
	}
	l.Barrels = append(l.Barrels, open[:n]...)
}
