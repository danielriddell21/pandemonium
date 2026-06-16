package world

// Floor heights are sculpted in quarter-wall steps: rooms sit on their own
// levels, corridors between them become staircases, the exit stands on a raised
// dais, and the occasional ledge is served by a lift platform. Every adjacent
// pair of walkable tiles ends up within one step of each other (lift ledges
// excepted), so the whole layout stays traversable on foot.
const (
	// StepHeight is one quarter-step of floor height, in wall units.
	StepHeight = 0.25
	// MinHeadroom is the smallest ceiling-to-floor gap a walkable tile may have.
	MinHeadroom = 0.8

	// maxRoomRaise is the highest room level, in quarter-steps (3 -> +0.75).
	maxRoomRaise = 3
	// daisRaise lifts the exit tile this many quarter-steps above its room.
	daisRaise = 2
	// ledgeRaise lifts a bonus ledge this many quarter-steps above its
	// surroundings — deliberately beyond what a step can climb.
	ledgeRaise = 3
	// smoothingSweeps bounds the relaxation passes; layouts settle far sooner.
	smoothingSweeps = 64
)

// assignHeights sculpts the level's floor and ceiling heights. It runs after the
// layout, items and locks are final, so it can keep required routes walkable and
// reserve only spare dead ends for lift ledges.
func assignHeights(l *Level, g *rng, rooms []rect) {
	lv := roomLevels(l, g, rooms)
	spreadToCorridors(l, lv)
	smoothSteps(l, lv)
	raiseDais(l, lv)
	applyLevels(l, lv)
	assignCeilings(l, g, rooms)
	placeLiftLedge(l, g)
}

// roomLevels picks a quarter-step floor level for every room and stamps it on the
// room's cells. The spawn's room is anchored at the base level so runs always
// start on familiar ground.
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

// contains reports whether the rectangle covers the cell.
func (r rect) contains(c Coord) bool {
	return c.X >= r.x && c.X < r.x+r.w && c.Y >= r.y && c.Y < r.y+r.h
}

// spreadToCorridors floods room levels outward so corridor cells inherit the
// level of the nearest room, leaving any level seams mid-corridor for smoothing
// to turn into staircases.
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

// smoothSteps relaxes the levels until no two adjacent walkable cells differ by
// more than one step, lowering the higher side of each violation. Heights only
// ever decrease, so the sweep terminates.
func smoothSteps(l *Level, lv []int) {
	for range smoothingSweeps {
		changed := false
		for y := range l.Height {
			for x := range l.Width {
				c := Coord{X: x, Y: y}
				if !l.At(x, y).Walkable() {
					continue
				}
				ci := y*l.Width + x
				for _, n := range neighbors4(c) {
					if !l.InBounds(n.X, n.Y) || !l.At(n.X, n.Y).Walkable() {
						continue
					}
					ni := n.Y*l.Width + n.X
					if lv[ci] > lv[ni]+1 {
						lv[ci] = lv[ni] + 1
						changed = true
					}
				}
			}
		}
		if !changed {
			return
		}
	}
}

// raiseDais lifts the exit two steps above its smoothed level, then relaxes the
// surroundings upward so the platform is approached by single steps from every
// side. Raising only the lower half of each violating pair keeps every existing
// staircase intact, and since levels only increase toward the dais height, the
// sweep terminates.
func raiseDais(l *Level, lv []int) {
	e := l.Exit
	if !l.InBounds(e.X, e.Y) {
		return
	}
	lv[e.Y*l.Width+e.X] += daisRaise

	for range smoothingSweeps {
		changed := false
		for y := range l.Height {
			for x := range l.Width {
				if !l.At(x, y).Walkable() {
					continue
				}
				ci := y*l.Width + x
				for _, n := range neighbors4(Coord{X: x, Y: y}) {
					if !l.InBounds(n.X, n.Y) || !l.At(n.X, n.Y).Walkable() {
						continue
					}
					ni := n.Y*l.Width + n.X
					if lv[ci] < lv[ni]-1 {
						lv[ci] = lv[ni] - 1
						changed = true
					}
				}
			}
		}
		if !changed {
			return
		}
	}
}

// applyLevels converts quarter-step levels into floor heights.
func applyLevels(l *Level, lv []int) {
	for i, v := range lv {
		if l.Tiles[i].Walkable() {
			l.FloorH[i] = float64(v) * StepHeight
		}
	}
}

// placeLiftLedge turns one spare dead end into a raised bonus ledge served by a
// lift: the dead-end cell rises beyond step reach and holds a reward, and its
// neck becomes the platform that travels up to it. Dead ends already spent on
// doors, secrets or items are left alone, so no required route or pickup ever
// depends on the lift.
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
			d := Coord{X: x, Y: y}
			if l.At(x, y) != TileFloor || taken[d] || secret[d] {
				continue
			}
			nb := walkableNeighbors(l, d)
			if len(nb) != 1 { // ledges grow only from dead ends
				continue
			}
			neck := nb[0]
			if l.At(neck.X, neck.Y) != TileFloor || taken[neck] || neck == l.Spawn || neck == l.Exit {
				continue
			}
			if !g.chance(0.5) {
				continue
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
			return // at most one lift per level
		}
	}
}

// assignCeilings gives each room its own ceiling height (taller halls read more
// dramatic) and corridors a lower one, always preserving headroom above the
// sculpted floor.
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
