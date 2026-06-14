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
	// attackRange is how far ahead the player's strike reaches, in tiles.
	attackRange = 3.0
	// attackArcCos is the cosine of the half-angle within which a target must lie
	// (~25°), so the strike only hits what is roughly ahead.
	attackArcCos = 0.9

	// meleeHealth and rangedHealth are the demons' starting hit points.
	meleeHealth  = 60.0
	rangedHealth = 45.0
	// meleeDamage is the player's bare strike damage (a melee demon needs two).
	meleeDamage = 50.0
	// painDuration is how long a wounded demon staggers before resuming.
	painDuration = 0.2
	// deathDuration is how long the death animation plays before settling.
	deathDuration = 0.5
)

// attack strikes straight ahead, wounding the nearest living demon in range.
func (g *Game) attack() {
	if i := g.hitscan(attackRange, attackArcCos); i >= 0 {
		g.damageEntity(i, meleeDamage)
	}
}

// hitscan returns the index of the nearest living demon within maxRange, inside
// the facing arc, and in clear line of sight, or -1 if none.
func (g *Game) hitscan(maxRange, arcCos float64) int {
	dir := g.Player.Dir()
	best := -1
	bestD := math.Inf(1)
	for i := range g.Entities {
		e := g.Entities[i]
		if !e.Alive {
			continue
		}
		dx, dy := e.Pos.X-g.Player.Pos.X, e.Pos.Y-g.Player.Pos.Y
		d := math.Hypot(dx, dy)
		if d == 0 || d > maxRange {
			continue
		}
		if (dx/d)*dir.X+(dy/d)*dir.Y < arcCos {
			continue
		}
		if !losClear(g.World, g.Player.Pos, e.Pos) {
			continue
		}
		if d < bestD {
			bestD, best = d, i
		}
	}
	return best
}

// damageEntity applies damage to a demon, staggering it or killing it.
func (g *Game) damageEntity(i int, dmg float64) {
	e := &g.Entities[i]
	if e.State != Active {
		return
	}
	e.Health -= dmg
	if e.Health <= 0 {
		e.State = Dying
		e.Alive = false
		e.anim = 0
	} else {
		e.hurt = painDuration
	}
}

// die emits the death observation and respawns the player at the level spawn with
// full health, resetting the demons to their deterministic starting layout.
func (g *Game) die() {
	g.emit(Observation{Kind: ObsDeath, At: g.PlayerCell()})
	l := g.World.Level
	g.Player.Pos = Vec2{X: float64(l.Spawn.X) + 0.5, Y: float64(l.Spawn.Y) + 0.5}
	g.Player.Angle = facing(l.Spawn, l.Exit)
	g.Player.Health = MaxHealth
	g.Entities = spawnEntities(l)
	g.Projectiles = nil
	g.tracker.lastCell = l.Spawn
	g.tracker.started = true
}

// updateEntities advances demon behaviour for one step: corpses settle, dying
// demons play out their animation, and active demons within detectRadius and line
// of sight chase the player (staggering briefly when hurt), stopping at contact.
func (g *Game) updateEntities(dt float64) {
	pp := g.Player.Pos
	for i := range g.Entities {
		e := &g.Entities[i]
		switch e.State {
		case Dead:
			continue
		case Dying:
			e.anim += dt
			if e.anim >= deathDuration {
				e.State = Dead
			}
			continue
		}

		e.anim += dt
		if e.Kind == Ranged && e.fire > 0 {
			e.fire -= dt
		}
		if e.hurt > 0 {
			e.hurt -= dt
			continue
		}
		d := dist(e.Pos, pp)
		if d > detectRadius || d == 0 || !losClear(g.World, e.Pos, pp) {
			continue
		}
		if e.Kind == Ranged && e.fire <= 0 && d <= rangedFireRange {
			g.spawnProjectile(e.Pos, pp)
			e.fire = rangedFireCooldown
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
