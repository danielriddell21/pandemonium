package world

import "github.com/danielriddell21/crucible/worldgen"

func placeItems(l *Level, g *worldgen.RNG) {
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
		case i == 0 && g.Chance(0.25): // an occasional backpack, never more than one
			kind = ItemBackpack
		case i == 1 && g.Chance(0.4): // and sometimes a single powerup
			kind = rollPowerup(g)
		}
		l.Items = append(l.Items, Item{Kind: kind, At: c})
	}
}

func rollPowerup(g *worldgen.RNG) ItemKind {
	switch g.IntN(5) {
	case 0, 1:
		return ItemSoul
	case 2:
		return ItemBerserk
	case 3:
		return ItemRadSuit
	default:
		// The megasphere and invulnerability are the rarest finds.
		if g.Chance(0.5) {
			return ItemMega
		}
		return ItemInvuln
	}
}

func rollConsumable(g *worldgen.RNG) ItemKind {
	switch g.IntN(12) {
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

func freeFloors(l *Level) []Coord {
	taken := make(map[Coord]bool, len(l.Items))
	for _, it := range l.Items {
		taken[it.At] = true
	}
	var out []Coord
	for y := range l.H {
		for x := range l.W {
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

func shuffleCoords(g *worldgen.RNG, cs []Coord) {
	for i := len(cs) - 1; i > 0; i-- {
		j := g.IntN(i + 1)
		cs[i], cs[j] = cs[j], cs[i]
	}
}
