package sim

import "math"

const (
	// rangedFireRange is how far a ranged demon will lob projectiles, in tiles.
	rangedFireRange = 7.0
	// rangedFireCooldown is the time between a ranged demon's shots, in seconds.
	rangedFireCooldown = 2.0
	// projectileSpeed is a projectile's travel speed in tiles per second.
	projectileSpeed = 6.0
	// projectileDamage is the health a projectile removes on impact.
	projectileDamage = 12.0
	// projectileHitRadius is how close a projectile must get to hit a target.
	projectileHitRadius = 0.4
	// demonEye is how far above a demon's feet its projectiles launch from.
	demonEye = 0.4
)

// Projectile is an in-flight attack (e.g. a fireball). It flies in a straight 3D
// line; shooter is the index of the entity that fired it (-1 for none) so it
// never hits its owner and so infighting kills aren't credited to the player.
type Projectile struct {
	Pos     Vec2
	Z       float64 // height above the base floor, in wall units
	Vel     Vec2
	VelZ    float64
	Damage  float64
	shooter int
	Alive   bool
}

// spawnProjectile launches a projectile from the shooter's eye toward the
// target's eye, so demons on ledges can still hit a player below (and vice
// versa). shooter is the firing entity's index.
func (g *Game) spawnProjectile(shooter int, from Vec2, fromZ float64, target Vec2, targetZ float64) {
	dx, dy := target.X-from.X, target.Y-from.Y
	d := dist(from, target)
	if d == 0 {
		return
	}
	flight := d / projectileSpeed
	g.Projectiles = append(g.Projectiles, Projectile{
		Pos:     from,
		Z:       fromZ,
		Vel:     Vec2{X: dx / d * projectileSpeed, Y: dy / d * projectileSpeed},
		VelZ:    (targetZ - fromZ) / flight,
		Damage:  projectileDamage,
		shooter: shooter,
		Alive:   true,
	})
}

// advanceProjectiles moves projectiles, removing those that strike level geometry
// — a wall, a floor rising into their path, or a ceiling dipping below it — or a
// body. A projectile that crosses another demon wounds it (infighting, no player
// credit); one that reaches the player damages the player. Dead projectiles are
// compacted out.
func (g *Game) advanceProjectiles(dt float64) {
	kept := g.Projectiles[:0]
	for _, p := range g.Projectiles {
		if !p.Alive {
			continue
		}
		p.Pos.X += p.Vel.X * dt
		p.Pos.Y += p.Vel.Y * dt
		p.Z += p.VelZ * dt
		tx, ty := int(math.Floor(p.Pos.X)), int(math.Floor(p.Pos.Y))
		if g.World.Solid(tx, ty) ||
			p.Z <= g.World.FloorAt(tx, ty) || p.Z >= g.World.CeilAt(tx, ty) {
			continue // absorbed by the level
		}
		if j := g.projectileHitsDemon(p); j >= 0 {
			g.wound(j, p.Damage, false) // demon hit demon — infighting, no credit
			continue
		}
		if dist(p.Pos, g.Player.Pos) < projectileHitRadius &&
			p.Z > g.Player.Z && p.Z < g.Player.Z+1 {
			g.hurtPlayer(p.Damage)
			continue // struck the player
		}
		kept = append(kept, p)
	}
	g.Projectiles = kept
}

// projectileHitsDemon returns the index of a living demon the projectile is
// touching (excluding its own shooter), or -1. Height must overlap, so a shot
// sails harmlessly over a demon on a much lower floor.
func (g *Game) projectileHitsDemon(p Projectile) int {
	for j := range g.Entities {
		e := &g.Entities[j]
		if !e.Alive || j == p.shooter {
			continue
		}
		if dist(p.Pos, e.Pos) < projectileHitRadius && p.Z > e.Z && p.Z < e.Z+1 {
			return j
		}
	}
	return -1
}
