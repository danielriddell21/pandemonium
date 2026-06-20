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

	for i, c := range floors[:n] {
		kind := rollConsumable(g)
		switch {
		case i == 0 && g.chance(0.25): // an occasional backpack, never more than one
			kind = ItemBackpack
		case i == 1 && g.chance(0.4): // and sometimes a single powerup
			kind = rollPowerup(g)
		}
		l.Items = append(l.Items, Item{Kind: kind, At: c})
	}
}

// rollPowerup picks one of the DOOM-style powerups, weighted toward the more
// common spheres.
func rollPowerup(g *rng) ItemKind {
	switch g.intn(5) {
	case 0, 1:
		return ItemSoul
	case 2:
		return ItemBerserk
	case 3:
		return ItemRadSuit
	default:
		// The megasphere and invulnerability are the rarest finds.
		if g.chance(0.5) {
			return ItemMega
		}
		return ItemInvuln
	}
}

// rollConsumable picks a consumable kind with health and ammo common and armour
// and rockets rarer, roughly matching how often each turns up in a DOOM level.
func rollConsumable(g *rng) ItemKind {
	switch g.intn(12) {
	case 0, 1, 2:
		return ItemHealth
	case 3, 4, 5:
		return ItemBullets
	case 6, 7:
		return ItemShells
	case 8, 9:
		return ItemArmor
	default:
		return ItemRockets
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
