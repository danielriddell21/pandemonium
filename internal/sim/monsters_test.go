package sim

import "testing"

func TestSpawnIncludesPinkyAndBaron(t *testing.T) {
	seen := map[EntityKind]bool{}
	for seed := int64(0); seed < 20; seed++ {
		g := newTestGame(t, seed)
		for _, e := range g.Entities {
			seen[e.Kind] = true
		}
	}
	for _, k := range []EntityKind{Melee, Pinky, Ranged, Gunner, Baron} {
		if !seen[k] {
			t.Errorf("expected demon kind %d to appear across the sampled levels", k)
		}
	}
}

func TestPinkyRushesAndBaronLumbers(t *testing.T) {
	if !(demonSpeedFor(Pinky) > demonSpeed && demonSpeedFor(Baron) < demonSpeed) {
		t.Errorf("pinky should be faster and baron slower than the default: pinky=%v default=%v baron=%v",
			demonSpeedFor(Pinky), demonSpeed, demonSpeedFor(Baron))
	}
}

func TestBaronIsTankyAndFiresHeavyProjectiles(t *testing.T) {
	if baronHealth <= rangedHealth {
		t.Errorf("baron (%v) should out-tank an imp (%v)", baronHealth, rangedHealth)
	}
	if !ranges(Baron) {
		t.Error("baron should attack at range")
	}

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
		t.Skip("no clear placement for a baron")
	}
	g.Entities = []Entity{{Pos: spot, Z: g.Player.Z, Kind: Baron, State: Active, Health: baronHealth, Alive: true}}
	for range 90 {
		g.Tick(Input{}, 1.0/60.0)
		if len(g.Projectiles) > 0 {
			break
		}
	}
	if len(g.Projectiles) == 0 {
		t.Fatal("baron should have launched a fireball")
	}
	if g.Projectiles[0].Damage != baronProjDamage {
		t.Errorf("baron fireball damage = %v, want %v", g.Projectiles[0].Damage, baronProjDamage)
	}
}
