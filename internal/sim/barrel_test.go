package sim

import "testing"

func TestBarrelBurstDamagesNeighboursAndChains(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Pos = Vec2{X: 50, Y: 50} // out of the blast
	// Two adjacent barrels and a demon all within blast radius of the first.
	g.Entities = []Entity{
		{Pos: Vec2{X: 5.0, Y: 5.0}, Kind: Barrel, State: Active, Health: barrelHealth, Alive: true},
		{Pos: Vec2{X: 6.0, Y: 5.0}, Kind: Barrel, State: Active, Health: barrelHealth, Alive: true},
		{Pos: Vec2{X: 5.0, Y: 6.0}, Kind: Melee, State: Active, Health: meleeHealth, Alive: true},
	}
	g.killsTotal = countDemons(g.Entities)
	g.kills = 0

	g.damageEntity(0, barrelHealth) // pop the first barrel

	if g.Entities[0].State == Active {
		t.Fatal("shot barrel should have burst")
	}
	if g.Entities[1].State == Active {
		t.Error("blast should chain to the adjacent barrel")
	}
	if g.Entities[2].State == Active {
		t.Error("blast should kill the nearby demon (barrelDamage > meleeHealth)")
	}
	// The demon felled by the blast counts; the barrels never do.
	if g.kills != 1 {
		t.Errorf("kills = %d, want 1 (the demon only)", g.kills)
	}
}

func TestBarrelBurstHitsPlayer(t *testing.T) {
	g := newTestGame(t, 7)
	g.Entities = []Entity{{Pos: Vec2{X: g.Player.Pos.X + 1, Y: g.Player.Pos.Y}, Z: g.Player.Z, Kind: Barrel, State: Active, Health: barrelHealth, Alive: true}}
	before := g.Player.Health
	g.damageEntity(0, barrelHealth)
	if g.Player.Health >= before {
		t.Errorf("a barrel bursting next to the player should hurt them: %v -> %v", before, g.Player.Health)
	}
}

func TestBarrelDoesNotClawThePlayer(t *testing.T) {
	g := newTestGame(t, 7)
	// A barrel sitting on the player must not drain health like a demon would.
	g.Entities = []Entity{{Pos: g.Player.Pos, Z: g.Player.Z, Kind: Barrel, State: Active, Health: barrelHealth, Alive: true}}
	before := g.Player.Health
	for range 60 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if g.Player.Health != before {
		t.Errorf("an intact barrel must not damage the player by contact: %v -> %v", before, g.Player.Health)
	}
}

func TestLevelHasBarrels(t *testing.T) {
	g := newTestGame(t, 7)
	barrels := 0
	for _, e := range g.Entities {
		if e.Kind == Barrel {
			barrels++
		}
	}
	if barrels == 0 {
		t.Error("expected a generated level to scatter some barrels")
	}
}
