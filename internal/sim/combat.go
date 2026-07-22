package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	detectRadius = 6.0

	demonSpeed = 1.7

	contactRange = 0.6

	contactDamage = 30.0

	attackRange = 3.0

	attackArcCos = 0.9

	meleeHealth  = 60.0
	rangedHealth = 45.0
	gunnerHealth = 40.0
	pinkyHealth  = 55.0
	baronHealth  = 180.0

	pinkySpeed = 3.0
	baronSpeed = 1.2

	baronProjSpeed  = 4.0
	baronProjDamage = 30.0

	gunnerCooldown = 1.4
	gunnerDamage   = 9.0

	barrelHealth = 20.0
	barrelDamage = 60.0
	barrelRadius = 2.2

	meleeDamage = 50.0

	painDuration = 0.2

	deathDuration = 0.5

	walkFPS     = 6.0
	deathFrames = 3
)

func (g *Game) attack() {
	if i := g.hitscan(attackRange, attackArcCos); i >= 0 {
		g.damageEntity(i, meleeDamage)
	}
}

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

func (g *Game) damageEntity(i int, dmg float64) { g.wound(i, dmg, true) }

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

func (g *Game) updateEntities(dt float64) {
	pp := g.Player.Pos
	for i := range g.Entities {
		e := &g.Entities[i]
		if advanceDeath(e, dt) {
			continue
		}
		if e.Kind == Barrel {
			g.settleEntity(e, dt) // barrels just sit on the floor; no AI
			continue
		}
		g.updateDemon(e, i, pp, dt)
	}
}

func advanceDeath(e *Entity, dt float64) bool {
	switch e.State {
	case Dead:
		e.Frame = deathFrames - 1
		return true
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
		return true
	}
	return false
}

func (g *Game) updateDemon(e *Entity, i int, pp Vec2, dt float64) {
	e.anim += dt
	e.Frame = int(e.anim * walkFPS)
	g.settleEntity(e, dt)
	if e.fire > 0 { // shooters cool down between attacks
		e.fire -= dt
	}
	if e.hurt > 0 {
		e.hurt -= dt
		return
	}
	d := dist(e.Pos, pp)
	if d > detectRadius || d == 0 || !losClear(g.World, e.Pos, pp) {
		return
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

func (g *Game) settleEntity(e *Entity, dt float64) {
	floor := g.World.FloorAt(int(math.Floor(e.Pos.X)), int(math.Floor(e.Pos.Y)))
	switch {
	case e.Z < floor:
		e.Z = floor
	case e.Z > floor:
		e.Z = max(floor, e.Z-fallSpeed*dt)
	}
}

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
