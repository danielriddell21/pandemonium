package sim

const (
	// rangedFireRange is how far a ranged demon will lob projectiles, in tiles.
	rangedFireRange = 7.0
	// rangedFireCooldown is the time between a ranged demon's shots, in seconds.
	rangedFireCooldown = 2.0
	// projectileSpeed is a projectile's travel speed in tiles per second.
	projectileSpeed = 6.0
	// projectileDamage is the health a projectile removes on impact.
	projectileDamage = 12.0
	// projectileHitRadius is how close a projectile must get to hit the player.
	projectileHitRadius = 0.4
)

// Projectile is a demon's in-flight attack (e.g. a fireball).
type Projectile struct {
	Pos    Vec2
	Vel    Vec2
	Damage float64
	Alive  bool
}

// spawnProjectile launches a projectile from a demon toward the player.
func (g *Game) spawnProjectile(from, target Vec2) {
	dx, dy := target.X-from.X, target.Y-from.Y
	d := dist(from, target)
	if d == 0 {
		return
	}
	g.Projectiles = append(g.Projectiles, Projectile{
		Pos:    from,
		Vel:    Vec2{X: dx / d * projectileSpeed, Y: dy / d * projectileSpeed},
		Damage: projectileDamage,
		Alive:  true,
	})
}

// advanceProjectiles moves projectiles, removing those that hit a wall and
// damaging the player on contact. Dead projectiles are compacted out.
func (g *Game) advanceProjectiles(dt float64) {
	kept := g.Projectiles[:0]
	for _, p := range g.Projectiles {
		if !p.Alive {
			continue
		}
		p.Pos.X += p.Vel.X * dt
		p.Pos.Y += p.Vel.Y * dt
		if g.World.Solid(int(p.Pos.X), int(p.Pos.Y)) {
			continue // absorbed by a wall
		}
		if dist(p.Pos, g.Player.Pos) < projectileHitRadius {
			g.hurtPlayer(p.Damage)
			continue // struck the player
		}
		kept = append(kept, p)
	}
	g.Projectiles = kept
}
