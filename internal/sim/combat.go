package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

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

	// Per-kind starting hit points.
	meleeHealth  = 60.0
	rangedHealth = 45.0
	gunnerHealth = 40.0
	pinkyHealth  = 55.0
	baronHealth  = 180.0
	// pinkySpeed and baronSpeed bracket the default demonSpeed.
	pinkySpeed = 3.0
	baronSpeed = 1.2
	// Baron fireballs are slower but hit much harder than an imp's.
	baronProjSpeed  = 4.0
	baronProjDamage = 30.0
	// gunnerCooldown is the time between a gunner's instant shots; gunnerDamage is
	// what each lands. Hitscan: no projectile, but it must have line of sight.
	gunnerCooldown = 1.4
	gunnerDamage   = 9.0

	// barrelHealth is how much punishment a barrel takes before bursting;
	// barrelDamage and barrelRadius size the blast that hits everything near it.
	barrelHealth = 20.0
	barrelDamage = 60.0
	barrelRadius = 2.2
	// meleeDamage is the player's bare strike damage (a melee demon needs two).
	meleeDamage = 50.0
	// painDuration is how long a wounded demon staggers before resuming.
	painDuration = 0.2
	// deathDuration is how long the death animation plays before settling.
	deathDuration = 0.5
	// walkFPS is the demon walk-cycle rate; deathFrames is the death sequence length.
	walkFPS     = 6.0
	deathFrames = 3
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

// damageEntity applies player-dealt damage to a demon, staggering or killing it
// and crediting the kill to the player.
func (g *Game) damageEntity(i int, dmg float64) { g.wound(i, dmg, true) }

// wound applies damage to a demon. credit marks a kill as the player's (counted
// and observed); infighting damage between demons passes false so a demon felling
// another never inflates the player's tally.
func (g *Game) wound(i int, dmg float64, credit bool) {
	e := &g.Entities[i]
	if e.State != Active {
		return
	}
	e.Health -= dmg
	if e.Health <= 0 {
		e.State = Dying
		e.Alive = false
		e.anim = 0
		if e.Kind == Barrel {
			g.explode(e.Pos, e.Z, barrelRadius, barrelDamage) // burst, catching everything nearby
			return
		}
		if credit {
			g.kills++
			g.emit(Observation{Kind: ObsKill, At: g.PlayerCell()})
		}
	} else {
		e.hurt = painDuration
	}
}

// explode applies a blast at center: radius damage to the player and to every
// active entity within radius. Used by both barrels and rockets. Barrels caught
// in the blast burst in turn, and because a bursting barrel is no longer Active
// the chain terminates on its own.
func (g *Game) explode(center Vec2, z, radius, dmg float64) {
	if dist(center, g.Player.Pos) <= radius && math.Abs(z-g.Player.Z) < world.MinHeadroom {
		g.faceHurt(center)
		g.hurtPlayer(dmg)
	}
	for i := range g.Entities {
		e := &g.Entities[i]
		if e.State != Active { // the bursting barrel is already Dying, so it's skipped
			continue
		}
		if dist(e.Pos, center) <= radius {
			g.wound(i, dmg, true) // blast kills count for the player
		}
	}
}

// die emits the death observation and respawns the player at the level spawn with
// full health, resetting the demons and pickups to their deterministic starting
// layout so a death is a clean restart of the level rather than a soft lock.
func (g *Game) die() {
	g.emit(Observation{Kind: ObsDeath, At: g.PlayerCell()})
	l := g.World.Level
	g.Player.Pos = Vec2{X: float64(l.Spawn.X) + 0.5, Y: float64(l.Spawn.Y) + 0.5}
	g.Player.Z = l.Floor(l.Spawn.X, l.Spawn.Y)
	g.viewZ = g.Player.Z
	g.Player.Angle = facing(l.Spawn, l.Exit)
	g.Player.Health = MaxHealth
	g.Entities = g.spawnEntities()
	g.Projectiles = nil
	g.Items = newItems(l) // pickups return with the demons, so ammo can be recovered
	g.kills = 0           // the demons are back; the kill tally restarts with them
	g.items = 0
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
			e.Frame = deathFrames - 1
			continue
		case Dying:
			e.anim += dt
			if f := int(e.anim / deathDuration * deathFrames); f < deathFrames {
				e.Frame = f
			} else {
				e.Frame = deathFrames - 1
			}
			if e.anim >= deathDuration {
				e.State = Dead
			}
			continue
		}

		if e.Kind == Barrel {
			g.settleEntity(e, dt) // barrels just sit on the floor; no AI
			continue
		}

		e.anim += dt
		e.Frame = int(e.anim * walkFPS)
		g.settleEntity(e, dt)
		if e.fire > 0 { // shooters cool down between attacks
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
		if ranges(e.Kind) && e.fire <= 0 && d <= rangedFireRange {
			speed, dmg := float64(projectileSpeed), float64(projectileDamage)
			if e.Kind == Baron {
				speed, dmg = baronProjSpeed, baronProjDamage
			}
			g.spawnProjectile(i, e.Pos, e.Z+demonEye, pp, g.Player.Z+eyeHeight, speed, dmg)
			e.fire = rangedFireCooldown
		}
		if e.Kind == Gunner && e.fire <= 0 {
			g.faceHurt(e.Pos)
			g.hurtPlayer(gunnerDamage) // instant hitscan — already has line of sight
			e.fire = gunnerCooldown
		}
		if d > contactRange {
			speed := demonSpeedFor(e.Kind)
			ux, uy := (pp.X-e.Pos.X)/d, (pp.Y-e.Pos.Y)/d
			e.Pos = resolveMove(g.World, e.Pos, e.Z, ux*speed*dt, uy*speed*dt)
		}
	}
}

// settleEntity resolves a demon's height against the floor underfoot, mirroring
// the player's step-up and fall behaviour (including riding lifts).
func (g *Game) settleEntity(e *Entity, dt float64) {
	floor := g.World.FloorAt(int(math.Floor(e.Pos.X)), int(math.Floor(e.Pos.Y)))
	switch {
	case e.Z < floor:
		e.Z = floor
	case e.Z > floor:
		e.Z = math.Max(floor, e.Z-fallSpeed*dt)
	}
}

// applyContactDamage drains the player's health while any living demon is in
// contact, clamping at zero. A demon on a ledge well above (or below) the player
// cannot claw across the height difference.
func (g *Game) applyContactDamage(dt float64) {
	touching := false
	var from Vec2
	for _, e := range g.Entities {
		if e.Kind == Barrel {
			continue // a barrel you brush past doesn't claw you
		}
		if e.Alive && dist(e.Pos, g.Player.Pos) < contactRange &&
			math.Abs(e.Z-g.Player.Z) < world.MinHeadroom {
			touching, from = true, e.Pos
			break
		}
	}
	if touching {
		g.faceHurt(from)
		g.hurtPlayer(contactDamage * dt)
	}
}

// applyHazard drains the player's health while they stand on a damaging floor
// tile (and are actually on the ground, not stepping over it).
func (g *Game) applyHazard(dt float64) {
	c := g.PlayerCell()
	rate := g.World.HazardAt(c.X, c.Y)
	if rate <= 0 {
		return
	}
	// A radiation suit shrugs off slime, but lava burns through it.
	if g.Player.RadSuited() && g.World.HazardKindAt(c.X, c.Y) == world.HazardNukage {
		return
	}
	if g.Player.Z-g.World.FloorAt(c.X, c.Y) > world.MinHeadroom {
		return // up on something above the hazard, not wading in it
	}
	g.hurtPlayer(rate * dt)
}

// faceHurt turns the status-bar face toward the source of a hit: ahead if it's
// roughly in front, otherwise to whichever side it came from.
func (g *Game) faceHurt(src Vec2) {
	dir := g.Player.Dir()
	tx, ty := src.X-g.Player.Pos.X, src.Y-g.Player.Pos.Y
	d := math.Hypot(tx, ty)
	if d == 0 {
		return
	}
	fwd := (tx*dir.X + ty*dir.Y) / d
	right := (tx*(-dir.Y) + ty*dir.X) / d
	g.Player.hurtTTL = hurtFaceDuration
	switch {
	case fwd > 0.5:
		g.Player.hurtDir = 0
	case right > 0:
		g.Player.hurtDir = 1
	default:
		g.Player.hurtDir = -1
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
