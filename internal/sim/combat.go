package sim

import "math"

const (
	// detectRadius is how close (in tiles) a demon must be, with clear line of
	// sight, before it starts chasing the player.
	detectRadius = 6.0
	// demonSpeed is a demon's chase speed in tiles per second (below the player's).
	demonSpeed = 1.7
	// contactRange is the distance at which a demon touches the player.
	contactRange = 0.6
	// contactDamage is health lost per second while a demon is in contact.
	contactDamage = 30.0
)

// updateEntities advances demon behaviour for one step: any living demon within
// detectRadius and line of sight moves toward the player, stopping at contact.
func (g *Game) updateEntities(dt float64) {
	pp := g.Player.Pos
	for i := range g.Entities {
		e := &g.Entities[i]
		if !e.Alive {
			continue
		}
		d := dist(e.Pos, pp)
		if d > detectRadius || d == 0 || !losClear(g.World, e.Pos, pp) {
			continue
		}
		if d > contactRange {
			ux, uy := (pp.X-e.Pos.X)/d, (pp.Y-e.Pos.Y)/d
			e.Pos = resolveMove(g.World, e.Pos, ux*demonSpeed*dt, uy*demonSpeed*dt)
		}
	}
}

// applyContactDamage drains the player's health while any living demon is in
// contact, clamping at zero.
func (g *Game) applyContactDamage(dt float64) {
	touching := false
	for _, e := range g.Entities {
		if e.Alive && dist(e.Pos, g.Player.Pos) < contactRange {
			touching = true
			break
		}
	}
	if touching {
		g.Player.Health -= contactDamage * dt
	}
	if g.Player.Health < 0 {
		g.Player.Health = 0
	}
}

// losClear reports whether the straight segment a→b crosses no solid tile.
func losClear(w *World, a, b Vec2) bool {
	steps := int(dist(a, b)/0.1) + 1
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := a.X + (b.X-a.X)*t
		y := a.Y + (b.Y-a.Y)*t
		if w.Solid(int(math.Floor(x)), int(math.Floor(y))) {
			return false
		}
	}
	return true
}

func dist(a, b Vec2) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}
