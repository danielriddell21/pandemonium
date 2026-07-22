package sim

import "math"

const (
	rangedFireRange = 7.0

	rangedFireCooldown = 2.0

	projectileSpeed = 6.0

	projectileDamage = 12.0

	projectileHitRadius = 0.4

	demonEye = 0.4

	playerShooter = -1

	rocketSpeed  = 8.0
	rocketRadius = 2.5
)

type Projectile struct {
	Pos     Vec2
	Z       float64
	Vel     Vec2
	VelZ    float64
	Damage  float64
	Splash  bool
	shooter int
	Alive   bool
}

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

func (g *Game) detonate(p Projectile) {
	if p.Splash {
		g.explode(p.Pos, p.Z, rocketRadius, p.Damage)
	}
}

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
