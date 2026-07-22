package world

func placeBarrels(l *Level, g *rng) {
	floors := freeFloors(l) // already excludes spawn area, exit and item cells
	shuffleCoords(g, floors)

	n := len(floors) / 30
	if n < 1 {
		n = 1
	}
	if n > 6 {
		n = 6
	}
	if n > len(floors) {
		n = len(floors)
	}
	l.Barrels = append(l.Barrels, floors[:n]...)
}
