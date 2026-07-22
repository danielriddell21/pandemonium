package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestRangedDemonFiresProjectile(t *testing.T) {
	g := newTestGame(t, 7)
	p := g.Player.Pos
	var spot Vec2
	found := false
	for r := 2.5; r <= 6; r += 0.5 {
		c := Vec2{X: p.X + r, Y: p.Y}
		if !g.World.Solid(int(c.X), int(c.Y)) && losClear(g.World, c, p) {
			spot, found = c, true
			break
		}
	}
	if !found {
		t.Skip("no clear placement for a ranged demon")
	}
	g.Entities = []Entity{{Pos: spot, Kind: Ranged, State: Active, Health: rangedHealth, Alive: true}}

	maxProj := 0
	for range 90 {
		g.Tick(Input{}, 1.0/60.0)
		if len(g.Projectiles) > maxProj {
			maxProj = len(g.Projectiles)
		}
	}
	if maxProj == 0 {
		t.Error("ranged demon should have fired at least one projectile")
	}
}

func TestProjectileStopsAtWall(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 5})
	if err != nil {
		t.Fatal(err)
	}
	g := New(l)
	g.Entities = nil

	// Find a floor cell with a solid neighbour and aim a projectile into it.
	var from, vel Vec2
	placed := false
	dirs := []world.Coord{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}}
	for y := 0; y < l.H && !placed; y++ {
		for x := 0; x < l.W && !placed; x++ {
			if l.At(x, y) != world.TileFloor {
				continue
			}
			for _, d := range dirs {
				if g.World.Solid(x+d.X, y+d.Y) {
					from = Vec2{X: float64(x) + 0.5, Y: float64(y) + 0.5}
					vel = Vec2{X: float64(d.X) * projectileSpeed, Y: float64(d.Y) * projectileSpeed}
					placed = true
					break
				}
			}
		}
	}
	if !placed {
		t.Skip("no wall-adjacent floor cell found")
	}
	g.Player.Pos = Vec2{X: -100, Y: -100} // keep the player clear
	z := g.World.FloorAt(int(from.X), int(from.Y)) + demonEye
	g.Projectiles = []Projectile{{Pos: from, Z: z, Vel: vel, Damage: projectileDamage, Alive: true}}

	for range 40 {
		g.advanceProjectiles(1.0 / 60.0)
	}
	if len(g.Projectiles) != 0 {
		t.Errorf("projectile should have been absorbed by the wall, %d remain", len(g.Projectiles))
	}
}

func TestProjectileDamagesPlayer(t *testing.T) {
	g := newTestGame(t, 7)
	g.Entities = nil
	p := g.Player.Pos
	g.Projectiles = []Projectile{{
		Pos:    Vec2{X: p.X + 0.6, Y: p.Y},
		Z:      g.Player.Z + eyeHeight,
		Vel:    Vec2{X: -projectileSpeed, Y: 0}, // straight at the player
		Damage: projectileDamage,
		Alive:  true,
	}}
	before := g.Player.Health
	for range 20 {
		g.advanceProjectiles(1.0 / 60.0)
	}
	if g.Player.Health >= before {
		t.Errorf("projectile should damage the player: %v -> %v", before, g.Player.Health)
	}
	if len(g.Projectiles) != 0 {
		t.Error("projectile should be consumed on hitting the player")
	}
}
