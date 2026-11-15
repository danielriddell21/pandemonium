package world

import "github.com/danielriddell21/crucible/worldgen"

func placeBarrels(l *Level, g *worldgen.RNG) {
	floors := freeFloors(l) // already excludes spawn area, exit and item cells

	// Barrels are solid, so one may only stand where it can be walked around
	// on every side: never in a corridor, a doorway or the mouth of one, where
	// it would wall off the way through.
	open := floors[:0]
	for _, c := range floors {
		if ringOpen(l, c) {
			open = append(open, c)
		}
	}
	shuffleCoords(g, open)

	n := min(max(len(open)/20, 1), 6, len(open))
	l.Barrels = append(l.Barrels, open[:n]...)
}

func ringOpen(l *Level, c Coord) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if !l.At(c.X+dx, c.Y+dy).Walkable() {
				return false
			}
		}
	}
	return true
}
