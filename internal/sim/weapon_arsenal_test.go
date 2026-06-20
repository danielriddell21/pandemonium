package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestChaingunSpendsBulletsRapidly(t *testing.T) {
	g := newTestGame(t, 7)
	g.Entities = nil
	g.Player.Weapon = Chaingun
	g.Player.Bullets = 5
	if !g.fire() || g.Player.Bullets != 4 {
		t.Fatalf("chaingun should spend one bullet per shot, have %d", g.Player.Bullets)
	}
	g.Player.Bullets = 0
	if g.fire() {
		t.Error("chaingun should not fire with no bullets")
	}
}

func TestRocketSpawnsProjectileAndSplashes(t *testing.T) {
	g := newTestGame(t, 7)
	g.Entities = nil
	g.Player.Weapon = RocketLauncher
	g.Player.Rockets = 2
	g.Player.Angle = 0

	if !g.fire() {
		t.Fatal("rocket launcher should fire with ammo")
	}
	if len(g.Projectiles) != 1 || !g.Projectiles[0].Splash {
		t.Fatalf("firing a rocket should spawn one splash projectile, got %+v", g.Projectiles)
	}
	if g.Player.Rockets != 1 {
		t.Errorf("rocket should spend one rocket, have %d", g.Player.Rockets)
	}
}

func TestRocketBlastKillsClusteredDemons(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Pos = Vec2{X: 5.5, Y: 5.5}
	g.Player.Z = 0
	g.Player.Angle = 0 // +X
	g.Player.Weapon = RocketLauncher
	g.Player.Rockets = 1
	// Two demons clustered a few tiles ahead, within the blast radius of impact.
	g.Entities = []Entity{
		{Pos: Vec2{X: 8.4, Y: 5.5}, Z: 0, Kind: Melee, State: Active, Health: meleeHealth, Alive: true},
		{Pos: Vec2{X: 8.6, Y: 6.0}, Z: 0, Kind: Melee, State: Active, Health: meleeHealth, Alive: true},
	}
	g.killsTotal = 2
	g.kills = 0
	g.fire()
	for range 120 {
		g.advanceProjectiles(1.0 / 60.0)
		if len(g.Projectiles) == 0 {
			break
		}
	}
	if g.Entities[0].State == Active || g.Entities[1].State == Active {
		t.Errorf("rocket blast should fell both clustered demons: %v %v", g.Entities[0].State, g.Entities[1].State)
	}
}

func TestAmmoPickupClampsToCapAndBackpackRaisesIt(t *testing.T) {
	g := newTestGame(t, 7)
	g.Entities = nil
	g.Player.Bullets = baseMaxBullets - 5
	g.applyPickup(world.ItemBullets) // +20 but capped
	if g.Player.Bullets != baseMaxBullets {
		t.Errorf("bullets should cap at %d, got %d", baseMaxBullets, g.Player.Bullets)
	}
	g.applyPickup(world.ItemBackpack)
	if !g.Player.Backpack || g.Player.MaxBullets() != baseMaxBullets*2 {
		t.Errorf("backpack should double the cap: backpack=%v max=%d", g.Player.Backpack, g.Player.MaxBullets())
	}
	g.applyPickup(world.ItemBullets) // now there's room again
	if g.Player.Bullets <= baseMaxBullets {
		t.Errorf("after backpack, bullets should rise above the base cap, got %d", g.Player.Bullets)
	}
}

func TestWheelAndSlotsCoverAllWeapons(t *testing.T) {
	g := newTestGame(t, 7)
	for slot, want := range map[int]WeaponKind{1: Fists, 2: Pistol, 3: Shotgun, 4: Chaingun, 5: RocketLauncher} {
		g.switchWeapon(slot)
		if g.Player.Weapon != want {
			t.Errorf("slot %d selected %v, want %v", slot, g.Player.Weapon, want)
		}
	}
}
