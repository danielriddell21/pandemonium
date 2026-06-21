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

	// playerShooter marks a projectile fired by the player (a rocket): it can hit
	// any demon and never registers a direct hit on the player, only splash.
	playerShooter = -1
	// rocketSpeed is how fast a player rocket travels; rocketRadius is its blast.
	rocketSpeed  = 8.0
	rocketRadius = 2.5
)

// Projectile is an in-flight attack (e.g. a fireball or a rocket). It flies in a
// straight 3D line; shooter is the index of the entity that fired it
// (playerShooter for the player) so it never hits its owner and so infighting
// kills aren't credited to the player. Splash projectiles burst on impact.
type Projectile struct {
	Pos     Vec2
	Z       float64 // height above the base floor, in wall units
	Vel     Vec2
	VelZ    float64
	Damage  float64
	Splash  bool // bursts for radius damage on impact (a rocket)
	shooter int
	Alive   bool
}

// spawnPlayerRocket launches a rocket from the player's eye along their facing.
func (g *Game) spawnPlayerRocket(dmg float64) {
	dir := g.Player.Dir()
	g.Projectiles = append(g.Projectiles, Projectile{
		Pos:     g.Player.Pos,
		Z:       g.Player.Z + eyeHeight,
		Vel:     Vec2{X: dir.X * rocketSpeed, Y: dir.Y * rocketSpeed},
		Damage:  dmg,
		Splash:  true,
		shooter: playerShooter,
		Alive:   true,
	})
}

// spawnProjectile launches a projectile from the shooter's eye toward the
// target's eye at the given speed and damage, so demons on ledges can still hit a
// player below (and vice versa). shooter is the firing entity's index.
func (g *Game) spawnProjectile(shooter int, from Vec2, fromZ float64, target Vec2, targetZ, speed, dmg float64) {
	dx, dy := target.X-from.X, target.Y-from.Y
	d := dist(from, target)
	if d == 0 {
		return
	}
	flight := d / speed
	g.Projectiles = append(g.Projectiles, Projectile{
		Pos:     from,
		Z:       fromZ,
		Vel:     Vec2{X: dx / d * speed, Y: dy / d * speed},
		VelZ:    (targetZ - fromZ) / flight,
		Damage:  dmg,
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
			g.detonate(p) // absorbed by the level (rockets burst here)
			continue
		}
		if j := g.projectileHitsDemon(p); j >= 0 {
			if p.Splash {
				g.detonate(p)
			} else {
				g.wound(j, p.Damage, false) // demon hit demon — infighting, no credit
			}
			continue
		}
		// Demon projectiles strike the player directly; player rockets only splash.
		if p.shooter != playerShooter && dist(p.Pos, g.Player.Pos) < projectileHitRadius &&
			p.Z > g.Player.Z && p.Z < g.Player.Z+1 {
			g.faceHurt(p.Pos)
			g.hurtPlayer(p.Damage)
			continue
		}
		kept = append(kept, p)
	}
	g.Projectiles = kept
}

// detonate bursts a splash projectile at its current position; non-splash
// projectiles simply vanish.
func (g *Game) detonate(p Projectile) {
	if p.Splash {
		g.explode(p.Pos, p.Z, rocketRadius, p.Damage)
	}
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
