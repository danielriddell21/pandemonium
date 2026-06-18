package sim

import "testing"

func TestGunnerHitsInstantlyOnCooldown(t *testing.T) {
	g := newTestGame(t, 7)
	p := g.Player.Pos
	// Place a gunner a few tiles ahead with clear line of sight.
	var spot Vec2
	found := false
	for r := 2.0; r <= 5; r += 0.5 {
		c := Vec2{X: p.X + r, Y: p.Y}
		if !g.World.Solid(int(c.X), int(c.Y)) && losClear(g.World, c, p) {
			spot, found = c, true
			break
		}
	}
	if !found {
		t.Skip("no clear gunner placement on this level")
	}
	g.Entities = []Entity{{Pos: spot, Z: g.Player.Z, Kind: Gunner, State: Active, Health: gunnerHealth, Alive: true}}

	before := g.Player.Health
	g.Tick(Input{}, 1.0/60.0) // first tick: fire <= 0, so it shoots immediately
	if g.Player.Health >= before {
		t.Fatalf("gunner should land an instant hit: %v -> %v", before, g.Player.Health)
	}
	if len(g.Projectiles) != 0 {
		t.Error("gunner is hitscan and must not spawn a projectile")
	}

	// It must respect the cooldown: no second hit on the very next tick.
	afterFirst := g.Player.Health
	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Health != afterFirst {
		t.Errorf("gunner fired again before its cooldown elapsed: %v -> %v", afterFirst, g.Player.Health)
	}
}

func TestGunnerNeedsLineOfSight(t *testing.T) {
	g := newTestGame(t, 7)
	// A gunner walled off from the player (corner cell) can't shoot.
	g.Entities = []Entity{{Pos: Vec2{X: 0.5, Y: 0.5}, Kind: Gunner, State: Active, Health: gunnerHealth, Alive: true}}
	before := g.Player.Health
	for range 120 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if g.Player.Health != before {
		t.Errorf("gunner with no line of sight should not hit: %v -> %v", before, g.Player.Health)
	}
}

func TestSpawnIncludesAllDemonKinds(t *testing.T) {
	// A reasonably large level should produce melee, ranged and gunner demons.
	g := newTestGame(t, 7)
	seen := map[EntityKind]bool{}
	for _, e := range g.Entities {
		seen[e.Kind] = true
	}
	for _, k := range []EntityKind{Melee, Ranged, Gunner} {
		if !seen[k] {
			t.Errorf("expected at least one demon of kind %d in a large level", k)
		}
	}
}
