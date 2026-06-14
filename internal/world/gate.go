package world

// placeKeyGate locks the route to the exit behind a keycard. It finds a narrow
// corridor cell on the spawn→exit path that, if sealed, truly cuts the exit off,
// turns it into a locked door, and drops the matching key somewhere the player
// can reach without first crossing that door. The BSP layout's corridors are a
// tree, so such cut points exist on every non-trivial level. Placement is
// best-effort: a candidate that offers no valid spot is simply left ungated, and
// Generate's reachability guarantee is the backstop.
func placeKeyGate(l *Level, _ *rng) {
	path := pathToExit(l)
	if len(path) == 0 {
		return
	}

	// Walk the path from the exit end inward: a lock nearer the exit leaves the
	// larger spawn-side region free to hold the key.
	for i := len(path) - 1; i >= 0; i-- {
		neck := path[i]
		if !isNeck(l, neck) {
			continue
		}
		blocked := lockedSolid(l, neck)
		// The neck must genuinely sever the exit (no way around it) and still
		// leave a reachable cell for the key.
		if reachable(l, l.Spawn, l.Exit, blocked) {
			continue
		}
		keyCell, ok := farthestReachableFloor(l, blocked)
		if !ok {
			continue
		}
		if l.Locks == nil {
			l.Locks = make(map[Coord]ItemKind)
		}
		l.set(neck.X, neck.Y, TileDoor)
		l.Locks[neck] = ItemKeyRed
		l.Items = append(l.Items, Item{Kind: ItemKeyRed, At: keyCell})
		return
	}
}

// lockedSolid treats walls and the given cell as solid while leaving every door
// passable — the world as seen by a player who lacks the key for that cell.
func lockedSolid(l *Level, lock Coord) solidFn {
	return func(c Coord) bool { return l.At(c.X, c.Y) == TileWall || c == lock }
}

// pathToExit reconstructs the spawn→exit route over the doors-open graph, or nil
// if the exit is unreachable.
func pathToExit(l *Level) []Coord {
	solid := blocksWalls(l)
	prev := make([]Coord, l.Width*l.Height)
	seen := make([]bool, l.Width*l.Height)
	queue := []Coord{l.Spawn}
	seen[l.Spawn.Y*l.Width+l.Spawn.X] = true
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == l.Exit {
			return tracePath(prev, l, c)
		}
		for _, n := range neighbors4(c) {
			if !l.InBounds(n.X, n.Y) || solid(n) {
				continue
			}
			idx := n.Y*l.Width + n.X
			if seen[idx] {
				continue
			}
			seen[idx] = true
			prev[idx] = c
			queue = append(queue, n)
		}
	}
	return nil
}

// tracePath walks parent links from dst back to the spawn and returns the route
// in spawn→dst order.
func tracePath(prev []Coord, l *Level, dst Coord) []Coord {
	var rev []Coord
	for c := dst; ; c = prev[c.Y*l.Width+c.X] {
		rev = append(rev, c)
		if c == l.Spawn {
			break
		}
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

// isNeck reports whether c is a plain one-wide corridor cell (degree-2 floor, not
// part of an open room and not the spawn or exit) — the kind of cell that gates
// cleanly when sealed.
func isNeck(l *Level, c Coord) bool {
	if l.At(c.X, c.Y) != TileFloor || c == l.Spawn || c == l.Exit {
		return false
	}
	if inOpenBlock(l, c) {
		return false
	}
	return len(walkableNeighbors(l, c)) == 2
}

// farthestReachableFloor returns the floor cell, clear of the spawn area, that is
// furthest from the spawn while the predicate holds, so the key sits deep in the
// reachable region rather than next to the start.
func farthestReachableFloor(l *Level, solid solidFn) (Coord, bool) {
	dist := floodDist(l, l.Spawn, solid)
	best, bestD := Coord{}, -1
	for y := range l.Height {
		for x := range l.Width {
			if l.At(x, y) != TileFloor {
				continue
			}
			c := Coord{X: x, Y: y}
			if c == l.Spawn || c == l.Exit || cheby(c, l.Spawn) < 3 {
				continue
			}
			if d := dist[y*l.Width+x]; d > bestD {
				best, bestD = c, d
			}
		}
	}
	return best, bestD >= 0
}

// keysReachable reports whether every locked door's key can be reached from the
// spawn without first crossing that door. It backs the generation guarantee for
// gated levels.
func keysReachable(l *Level) bool {
	for lock, key := range l.Locks {
		cell, ok := keyCell(l, key)
		if !ok {
			return false
		}
		if !reachable(l, l.Spawn, cell, lockedSolid(l, lock)) {
			return false
		}
	}
	return true
}

// keyCell returns the location of the keycard item of the given kind.
func keyCell(l *Level, key ItemKind) (Coord, bool) {
	for _, it := range l.Items {
		if it.Kind == key {
			return it.At, true
		}
	}
	return Coord{}, false
}
