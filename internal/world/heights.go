package world

const (
	StepHeight = 0.25

	MinHeadroom = 0.8

	maxRoomRaise = 3

	daisRaise = 2

	ledgeRaise = 3

	smoothingSweeps = 64
)

func assignHeights(l *Level, g *rng, rooms []rect) {
	lv := roomLevels(l, g, rooms)
	spreadToCorridors(l, lv)
	smoothSteps(l, lv)
	raiseDais(l, lv)
	applyLevels(l, lv)
	assignCeilings(l, g, rooms)
	placeLiftLedge(l, g)
}

func roomLevels(l *Level, g *rng, rooms []rect) []int {
	lv := make([]int, l.Width*l.Height)
	for i := range lv {
		lv[i] = -1 // unset
	}
	for _, r := range rooms {
		level := 0
		if !r.contains(l.Spawn) && g.chance(0.35) {
			level = g.between(1, maxRoomRaise)
		}
		for y := r.y; y < r.y+r.h; y++ {
			for x := r.x; x < r.x+r.w; x++ {
				if l.At(x, y).Walkable() {
					lv[y*l.Width+x] = level
				}
			}
		}
	}
	return lv
}

func (r rect) contains(c Coord) bool {
	return c.X >= r.x && c.X < r.x+r.w && c.Y >= r.y && c.Y < r.y+r.h
}

func spreadToCorridors(l *Level, lv []int) {
	var queue []Coord
	for y := range l.Height {
		for x := range l.Width {
			if lv[y*l.Width+x] >= 0 {
				queue = append(queue, Coord{X: x, Y: y})
			}
		}
	}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, n := range neighbors4(c) {
			if !l.InBounds(n.X, n.Y) || !l.At(n.X, n.Y).Walkable() {
				continue
			}
			idx := n.Y*l.Width + n.X
			if lv[idx] >= 0 {
				continue
			}
			lv[idx] = lv[c.Y*l.Width+c.X]
			queue = append(queue, n)
		}
	}
	// Any walkable cell missed entirely (no room on the map edge cases) sits at
	// the base level.
	for i := range lv {
		if lv[i] < 0 {
			lv[i] = 0
		}
	}
}

func relaxLevels(l *Level, lv []int, adjust func(cur, neighbour int) (int, bool)) {
	for range smoothingSweeps {
		changed := false
		for y := range l.Height {
			for x := range l.Width {
				if relaxCell(l, lv, x, y, adjust) {
					changed = true
				}
			}
		}
		if !changed {
			return
		}
	}
}

func relaxCell(l *Level, lv []int, x, y int, adjust func(cur, neighbour int) (int, bool)) bool {
	if !l.At(x, y).Walkable() {
		return false
	}
	ci := y*l.Width + x
	changed := false
	for _, n := range neighbors4(Coord{X: x, Y: y}) {
		if !l.InBounds(n.X, n.Y) || !l.At(n.X, n.Y).Walkable() {
			continue
		}
		if nv, ok := adjust(lv[ci], lv[n.Y*l.Width+n.X]); ok {
			lv[ci] = nv
			changed = true
		}
	}
	return changed
}

func smoothSteps(l *Level, lv []int) {
	relaxLevels(l, lv, func(cur, neighbour int) (int, bool) {
		if cur > neighbour+1 {
			return neighbour + 1, true
		}
		return cur, false
	})
}

func raiseDais(l *Level, lv []int) {
	e := l.Exit
	if !l.InBounds(e.X, e.Y) {
		return
	}
	lv[e.Y*l.Width+e.X] += daisRaise

	relaxLevels(l, lv, func(cur, neighbour int) (int, bool) {
		if cur < neighbour-1 {
			return neighbour - 1, true
		}
		return cur, false
	})
}

func applyLevels(l *Level, lv []int) {
	for i, v := range lv {
		if l.Tiles[i].Walkable() {
			l.FloorH[i] = float64(v) * StepHeight
		}
	}
}

func placeLiftLedge(l *Level, g *rng) {
	taken := make(map[Coord]bool, len(l.Items))
	for _, it := range l.Items {
		taken[it.At] = true
	}
	secret := make(map[Coord]bool, len(l.Secrets))
	for _, s := range l.Secrets {
		secret[s] = true
	}

	for y := range l.Height {
		for x := range l.Width {
			if tryLiftLedge(l, g, taken, secret, Coord{X: x, Y: y}) {
				return // at most one lift per level
			}
		}
	}
}

func tryLiftLedge(l *Level, g *rng, taken, secret map[Coord]bool, d Coord) bool {
	if l.At(d.X, d.Y) != TileFloor || taken[d] || secret[d] {
		return false
	}
	nb := walkableNeighbors(l, d)
	if len(nb) != 1 { // ledges grow only from dead ends
		return false
	}
	neck := nb[0]
	if l.At(neck.X, neck.Y) != TileFloor || taken[neck] || neck == l.Spawn || neck == l.Exit {
		return false
	}
	if !g.chance(0.5) {
		return false
	}
	low := l.Floor(neck.X, neck.Y)
	high := low + float64(ledgeRaise)*StepHeight
	l.setFloor(d.X, d.Y, high)
	// Both the ledge and the lift shaft need headroom above the raised
	// platform, not just above their static floors.
	l.setCeil(d.X, d.Y, high+1)
	l.setCeil(neck.X, neck.Y, high+1)
	if l.Lifts == nil {
		l.Lifts = make(map[Coord]Lift)
	}
	l.Lifts[neck] = Lift{Low: low, High: high}
	l.Items = append(l.Items, Item{Kind: secretReward(g), At: d})
	return true
}

func assignCeilings(l *Level, g *rng, rooms []rect) {
	inRoom := make([]bool, l.Width*l.Height)
	for _, r := range rooms {
		headroom := 1.0 + 0.2*float64(g.intn(4)) // 1.0 .. 1.6
		for y := r.y; y < r.y+r.h; y++ {
			for x := r.x; x < r.x+r.w; x++ {
				if !l.At(x, y).Walkable() {
					continue
				}
				inRoom[y*l.Width+x] = true
				l.setCeil(x, y, l.Floor(x, y)+headroom)
			}
		}
	}
	const corridorHeadroom = 0.9
	for y := range l.Height {
		for x := range l.Width {
			i := y*l.Width + x
			if !l.Tiles[i].Walkable() || inRoom[i] {
				continue
			}
			l.setCeil(x, y, l.Floor(x, y)+corridorHeadroom)
		}
	}
}
