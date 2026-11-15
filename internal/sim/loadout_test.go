package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestFreshGameUsesTheStartingLoadout(t *testing.T) {
	g := openRoom(t, 20, 12)
	if got := g.Loadout(); got != startingLoadout() {
		t.Errorf("fresh loadout = %+v, want %+v", got, startingLoadout())
	}
	if g.Depth() != 0 {
		t.Errorf("fresh depth = %d", g.Depth())
	}
}

func TestLoadoutCarriesKitAndPatchesWounds(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Health, g.Player.Armor = 12, 70
	g.Player.Bullets, g.Player.Shells, g.Player.Rockets = 7, 3, 9
	g.Player.Backpack, g.Player.Berserk, g.Player.Weapon = true, true, RocketLauncher
	g.Player.Keys = map[world.ItemKind]bool{world.ItemKeyRed: true}
	g.Player.InvulnTTL = 5

	next := openRoom(t, 20, 12, WithLoadout(g.Loadout()), WithDepth(3))
	p := next.Player
	if p.Health != MaxHealth {
		t.Errorf("wounds should be patched between floors: health %v", p.Health)
	}
	if p.Armor != 70 || p.Bullets != 7 || p.Shells != 3 || p.Rockets != 9 || !p.Backpack || !p.Berserk || p.Weapon != RocketLauncher {
		t.Errorf("kit did not carry: %+v", p)
	}
	if p.HasKey(world.ItemKeyRed) || p.Invulnerable() {
		t.Error("keys and timed powerups stay on the floor they were found on")
	}
	if next.Depth() != 3 {
		t.Errorf("depth = %d, want 3", next.Depth())
	}
}

func TestOverhealCarries(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Health = 180
	if got := g.Loadout().Health; got != 180 {
		t.Errorf("overheal should carry: %v", got)
	}
}

func TestDepthNeverNegative(t *testing.T) {
	if got := openRoom(t, 20, 12, WithDepth(-4)).Depth(); got != 0 {
		t.Errorf("depth = %d, want 0", got)
	}
}

func TestDeeperLevelsHoldMoreDemons(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 40, Height: 28, Seed: 9})
	if err != nil {
		t.Fatal(err)
	}
	shallow, deep := entityCount(l, SkillNormal, 0), entityCount(l, SkillNormal, 8)
	if deep <= shallow {
		t.Errorf("depth 8 holds %d demons, depth 0 holds %d: want more", deep, shallow)
	}
	if got := entityCount(l, SkillNightmare, 1000); got > maxDemons {
		t.Errorf("demon count %d exceeds the cap %d", got, maxDemons)
	}
	if depthScale(0) != 1 || depthScale(1000) != maxDepthMul {
		t.Errorf("depthScale bounds wrong: %v, %v", depthScale(0), depthScale(1000))
	}
}

func TestWeaponReach(t *testing.T) {
	if rng, arc := Pistol.Reach(); rng != weapons[Pistol].rng || arc != weapons[Pistol].arcCos {
		t.Errorf("pistol reach = %v, %v", rng, arc)
	}
	if rng, arc := RocketLauncher.Reach(); rng != rocketReach || arc != rocketArcCos {
		t.Errorf("rocket reach = %v, %v", rng, arc)
	}
}

func TestFeedbackSignals(t *testing.T) {
	g := openRoom(t, 30, 12)
	g.Player.Pos, g.Player.Angle = Vec2{X: 4.5, Y: 5.5}, 0
	if g.Pain() != 0 || g.HitMarker() || g.Stride() != 0 {
		t.Fatal("a fresh game should be at rest")
	}

	g.hurtPlayer(20)
	if g.Pain() <= 0 {
		t.Fatal("damage should register as pain")
	}
	slime := openRoom(t, 30, 12)
	for range 30 {
		slime.hurtPlayer(0.1) // a pool's trickle, tick after tick
		slime.tickFeedback(1.0 / 60)
	}
	if slime.Pain() < painFloor-painFade/60 {
		t.Errorf("sustained light damage should stay visible: pain %v", slime.Pain())
	}
	g.hurtPlayer(1000)
	if g.Pain() != 1 {
		t.Errorf("pain should saturate at 1, got %v", g.Pain())
	}
	g.Player.Health = MaxHealth
	run(g, Input{}, 120)
	if g.Pain() != 0 {
		t.Errorf("pain should fade: %v", g.Pain())
	}

	g.Entities = []Entity{demonAt(7.5, 5.5, Baron)}
	g.Player.InvulnTTL = 100
	g.fire()
	if !g.HitMarker() {
		t.Error("a landed shot should raise the hit marker")
	}
	run(g, Input{}, 30)
	if g.HitMarker() {
		t.Error("the hit marker should clear")
	}

	g.Entities = nil
	run(g, Input{Forward: 1}, 30)
	walked := g.Stride()
	if walked <= 0 {
		t.Error("walking should advance the stride")
	}
	run(g, Input{}, 30)
	if g.Stride() != walked {
		t.Error("standing still must not advance the stride")
	}
}

func TestInfightingDoesNotRaiseTheHitMarker(t *testing.T) {
	g := openRoom(t, 30, 12)
	g.Entities = []Entity{demonAt(7.5, 5.5, Baron)}
	g.wound(0, 5, false)
	if g.HitMarker() {
		t.Error("a demon hurting a demon is not the player's hit")
	}
}
