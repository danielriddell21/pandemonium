package sim

import "testing"

func TestProjectileWoundsDemonWithoutCredit(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Pos = Vec2{X: 50, Y: 50} // keep the player well clear
	// A target demon sitting just ahead of an in-flight fireball fired by demon 0.
	target := Vec2{X: 5.5, Y: 5.5}
	g.Entities = []Entity{
		{Pos: Vec2{X: 3.5, Y: 5.5}, Z: 0, Kind: Ranged, State: Active, Health: rangedHealth, Alive: true},
		{Pos: target, Z: 0, Kind: Melee, State: Active, Health: meleeHealth, Alive: true},
	}
	g.kills = 0
	g.Projectiles = []Projectile{{
		Pos: Vec2{X: 5.2, Y: 5.5}, Z: 0.5,
		Vel: Vec2{X: projectileSpeed}, Damage: projectileDamage, shooter: 0, Alive: true,
	}}
	hpBefore := g.Entities[1].Health

	g.advanceProjectiles(1.0 / 60.0)

	if g.Entities[1].Health >= hpBefore {
		t.Errorf("fireball should wound the demon in its path: %v -> %v", hpBefore, g.Entities[1].Health)
	}
	if g.kills != 0 {
		t.Errorf("infighting damage must not credit the player: kills = %d", g.kills)
	}
	if len(g.Projectiles) != 0 {
		t.Error("projectile should be consumed when it wounds a demon")
	}
}

func TestProjectileSkipsItsOwnShooter(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Pos = Vec2{X: 50, Y: 50}
	g.Entities = []Entity{{Pos: Vec2{X: 5.5, Y: 5.5}, Z: 0, Kind: Ranged, State: Active, Health: rangedHealth, Alive: true}}
	g.Projectiles = []Projectile{{
		Pos: Vec2{X: 5.5, Y: 5.5}, Z: 0.5, // sitting right on the shooter
		Vel: Vec2{X: projectileSpeed}, Damage: projectileDamage, shooter: 0, Alive: true,
	}}
	hp := g.Entities[0].Health
	g.advanceProjectiles(1.0 / 600.0) // tiny step: stays near the shooter
	if g.Entities[0].Health != hp {
		t.Errorf("a projectile must never hit its own shooter: %v -> %v", hp, g.Entities[0].Health)
	}
}
