package world

import "github.com/danielriddell21/crucible/worldgen"

func annotate(l *Level) {
	g := worldgen.NewRNG(l.Seed ^ 0x5bd1e995)
	distExit := distanceField(l, l.Exit)

	markers := junctionMarkers(l, distExit)
	markers = append(markers, deadEndDoorMarkers(l, g)...)
	if m, ok := decoyExitMarker(l, g); ok {
		markers = append(markers, m)
	}
	l.Markers = markers
}

func junctionMarkers(l *Level, distExit []int) []Marker {
	var markers []Marker
	for y := range l.H {
		for x := range l.W {
			c := Coord{X: x, Y: y}
			if !l.At(x, y).Walkable() || c == l.Spawn || c == l.Exit {
				continue
			}
			nb := walkableNeighbors(l, c)
			if len(nb) < 3 || inOpenBlock(l, c) {
				continue
			}
			opt, best := c, -1
			for _, n := range nb {
				d := distExit[n.Y*l.W+n.X]
				if d >= 0 && (best < 0 || d < best) {
					best, opt = d, n
				}
			}
			markers = append(markers, Marker{
				Kind:     MarkerJunction,
				At:       c,
				Branches: nb,
				Optimal:  opt,
			})
		}
	}
	return markers
}

func deadEndDoorMarkers(l *Level, g *worldgen.RNG) []Marker {
	const maxDoors = 3
	var markers []Marker
	for y := range l.H {
		for x := range l.W {
			if len(markers) >= maxDoors {
				return markers
			}
			if m, ok := tryDeadEndDoor(l, g, Coord{X: x, Y: y}); ok {
				markers = append(markers, m)
			}
		}
	}
	return markers
}

func tryDeadEndDoor(l *Level, g *worldgen.RNG, d Coord) (Marker, bool) {
	if l.At(d.X, d.Y) != TileFloor {
		return Marker{}, false
	}
	nb := walkableNeighbors(l, d)
	if len(nb) != 1 { // not a dead end
		return Marker{}, false
	}
	neck := nb[0]
	if neck == l.Spawn || neck == l.Exit || l.At(neck.X, neck.Y) != TileFloor {
		return Marker{}, false
	}
	nnb := walkableNeighbors(l, neck)
	if len(nnb) != 2 { // neck must be a plain corridor cell
		return Marker{}, false
	}
	if !g.Chance(0.6) {
		return Marker{}, false
	}
	// The branch that is not the dead end is the way back out.
	optimal := nnb[0]
	if optimal == d {
		optimal = nnb[1]
	}
	l.Set(neck.X, neck.Y, TileDoor)
	// The cell tucked behind the door is a natural hidden room: half the time,
	// mark it a secret and stash a reward there.
	if g.Chance(0.5) {
		l.Secrets = append(l.Secrets, d)
		l.Items = append(l.Items, Item{Kind: secretReward(g), At: d})
	}
	return Marker{
		Kind:     MarkerDeadEndDoor,
		At:       neck,
		Branches: nnb,
		Optimal:  optimal,
	}, true
}

func secretReward(g *worldgen.RNG) ItemKind {
	switch g.IntN(5) {
	case 0, 1:
		return ItemArmor
	case 2, 3:
		return ItemShells
	default:
		return ItemHealth
	}
}

func decoyExitMarker(l *Level, g *worldgen.RNG) (Marker, bool) {
	for radius := 2; radius <= 5; radius++ {
		var ring []Coord
		for y := l.Exit.Y - radius; y <= l.Exit.Y+radius; y++ {
			for x := l.Exit.X - radius; x <= l.Exit.X+radius; x++ {
				c := Coord{X: x, Y: y}
				if cheby(c, l.Exit) != radius {
					continue
				}
				if l.At(x, y) == TileFloor && c != l.Spawn {
					ring = append(ring, c)
				}
			}
		}
		if len(ring) > 0 {
			c := ring[g.IntN(len(ring))]
			return Marker{Kind: MarkerDecoyExit, At: c}, true
		}
	}
	return Marker{}, false
}

func inOpenBlock(l *Level, c Coord) bool {
	for _, corner := range [4]Coord{
		{X: c.X - 1, Y: c.Y - 1}, {X: c.X, Y: c.Y - 1}, {X: c.X - 1, Y: c.Y}, {X: c.X, Y: c.Y},
	} {
		if l.At(corner.X, corner.Y).Walkable() &&
			l.At(corner.X+1, corner.Y).Walkable() &&
			l.At(corner.X, corner.Y+1).Walkable() &&
			l.At(corner.X+1, corner.Y+1).Walkable() {
			return true
		}
	}
	return false
}

func distanceField(l *Level, src Coord) []int {
	dist := make([]int, l.W*l.H)
	for i := range dist {
		dist[i] = -1
	}
	if !l.InBounds(src.X, src.Y) || !l.At(src.X, src.Y).Walkable() {
		return dist
	}
	dist[src.Y*l.W+src.X] = 0
	queue := []Coord{src}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		base := dist[c.Y*l.W+c.X]
		for _, n := range neighbors4(c) {
			if !l.InBounds(n.X, n.Y) || !l.At(n.X, n.Y).Walkable() {
				continue
			}
			idx := n.Y*l.W + n.X
			if dist[idx] != -1 {
				continue
			}
			dist[idx] = base + 1
			queue = append(queue, n)
		}
	}
	return dist
}

func walkableNeighbors(l *Level, c Coord) []Coord {
	var out []Coord
	for _, n := range neighbors4(c) {
		if l.InBounds(n.X, n.Y) && l.At(n.X, n.Y).Walkable() {
			out = append(out, n)
		}
	}
	return out
}

func cheby(a, b Coord) int {
	dx, dy := a.X-b.X, a.Y-b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx
	}
	return dy
}
