package sim

import (
	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	crowdGap     = 0.45
	scentRefresh = 30
)

type scent struct {
	dist  []int
	from  world.Coord
	age   int
	valid bool
}

func (g *Game) trackScent() {
	s := &g.scent
	s.age++
	cell := g.PlayerCell()
	if s.valid && cell == s.from && s.age < scentRefresh {
		return
	}
	s.from, s.age, s.valid = cell, 0, true
	g.floodScent(cell)
}

func (g *Game) floodScent(src world.Coord) {
	l := g.World.Level
	s := &g.scent
	if len(s.dist) != l.W*l.H {
		s.dist = make([]int, l.W*l.H)
	}
	for i := range s.dist {
		s.dist[i] = -1
	}
	if !l.InBounds(src.X, src.Y) {
		return
	}
	s.dist[src.Y*l.W+src.X] = 0
	queue := []world.Coord{src}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, n := range neighbours(c) {
			// The trail is walked toward the player, so the step that matters
			// is the one from the farther cell n into the nearer cell c.
			if !g.demonPassable(n) || s.dist[n.Y*l.W+n.X] >= 0 || !l.StepOK(n, c, world.MaxStep, world.MinHeadroom) {
				continue
			}
			s.dist[n.Y*l.W+n.X] = s.dist[c.Y*l.W+c.X] + 1
			queue = append(queue, n)
		}
	}
}

func (g *Game) demonPassable(c world.Coord) bool {
	l := g.World.Level
	if !l.InBounds(c.X, c.Y) || !l.At(c.X, c.Y).Walkable() {
		return false
	}
	if !g.World.IsDoor(c.X, c.Y) {
		return true
	}
	if _, locked := g.World.Lock(c.X, c.Y); locked {
		return g.World.Opened(c.X, c.Y) // a keycard door holds until the player opens it
	}
	return true
}

func (g *Game) scentAt(c world.Coord) int {
	l := g.World.Level
	if !g.scent.valid || !l.InBounds(c.X, c.Y) {
		return -1
	}
	return g.scent.dist[c.Y*l.W+c.X]
}

func (g *Game) stalk(e *Entity, i int, speed, dt float64) {
	here := cellOf(e.Pos)
	best, bestD := here, g.scentAt(here)
	if bestD <= 0 {
		return // no trail from here, or already in the player's cell
	}
	for _, n := range neighbours(here) {
		if d := g.scentAt(n); d >= 0 && d < bestD {
			best, bestD = n, d
		}
	}
	if best == here {
		return
	}
	if g.World.IsDoor(best.X, best.Y) && !g.World.Opened(best.X, best.Y) {
		g.World.ForceOpenDoor(best.X, best.Y) // demons on the hunt shoulder through doors
	}
	g.advance(e, i, centreOf(best), speed, dt)
}

func (g *Game) advance(e *Entity, i int, target Vec2, speed, dt float64) {
	d := dist(e.Pos, target)
	if d == 0 {
		return
	}
	step := min(speed*dt, d)
	next := g.step(e.Pos, e.Z, (target.X-e.Pos.X)/d*step, (target.Y-e.Pos.Y)/d*step)
	if g.crowded(i, next) {
		return
	}
	e.Pos = next
}

func (g *Game) crowded(i int, next Vec2) bool {
	// Demons queue rather than stack: one gives way to any other that is
	// already nearer the player and would be stood on.
	own := dist(next, g.Player.Pos)
	for j := range g.Entities {
		o := &g.Entities[j]
		if j == i || !o.Alive || o.Kind == Barrel {
			continue
		}
		if dist(o.Pos, next) < crowdGap && dist(o.Pos, g.Player.Pos) < own {
			return true
		}
	}
	return false
}

func neighbours(c world.Coord) [4]world.Coord {
	return [4]world.Coord{
		{X: c.X + 1, Y: c.Y},
		{X: c.X - 1, Y: c.Y},
		{X: c.X, Y: c.Y + 1},
		{X: c.X, Y: c.Y - 1},
	}
}
