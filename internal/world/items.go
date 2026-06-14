package world

// placeItems scatters consumable pickups (health, armour, ammo) across the
// level's floor, away from the spawn so the player has to explore to find them.
// Placement is driven by the level's RNG, so a seed always yields the same
// layout. Keycards and secret rewards are placed separately.
func placeItems(l *Level, g *rng) {
	floors := freeFloors(l)
	shuffleCoords(g, floors)

	n := len(floors) / 12
	if n < 2 {
		n = 2
	}
	if n > 12 {
		n = 12
	}
	if n > len(floors) {
		n = len(floors)
	}

	for _, c := range floors[:n] {
		l.Items = append(l.Items, Item{Kind: rollConsumable(g), At: c})
	}
}

// rollConsumable picks a consumable kind with health and ammo common and armour
// rarer, roughly matching how often each turns up in a DOOM level.
func rollConsumable(g *rng) ItemKind {
	switch g.intn(10) {
	case 0, 1, 2:
		return ItemHealth
	case 3, 4, 5:
		return ItemBullets
	case 6, 7:
		return ItemShells
	default:
		return ItemArmor
	}
}

// freeFloors returns the plain floor cells that are clear of the spawn area, the
// exit and any cell already holding an item, in row-major order.
func freeFloors(l *Level) []Coord {
	taken := make(map[Coord]bool, len(l.Items))
	for _, it := range l.Items {
		taken[it.At] = true
	}
	var out []Coord
	for y := range l.Height {
		for x := range l.Width {
			if l.At(x, y) != TileFloor {
				continue
			}
			c := Coord{X: x, Y: y}
			if c == l.Spawn || c == l.Exit || taken[c] || cheby(c, l.Spawn) < 3 {
				continue
			}
			out = append(out, c)
		}
	}
	return out
}

// shuffleCoords does an in-place deterministic Fisher-Yates shuffle using the
// level RNG, so callers can take the first N cells as a random selection.
func shuffleCoords(g *rng, cs []Coord) {
	for i := len(cs) - 1; i > 0; i-- {
		j := g.intn(i + 1)
		cs[i], cs[j] = cs[j], cs[i]
	}
}
