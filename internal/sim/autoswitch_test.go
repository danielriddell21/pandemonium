package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func autoSwitchGame(t *testing.T) *Game {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	g := New(l)
	g.Entities = nil // no targets needed; we only exercise ammo bookkeeping
	return g
}

func TestAutoSwitchOffEmptyWeapon(t *testing.T) {
	g := autoSwitchGame(t)
	g.Player.Weapon = Shotgun
	g.Player.Shells = 0   // shotgun dry
	g.Player.Bullets = 10 // chaingun/pistol still usable
	g.autoSwitchIfEmpty()
	if g.Player.Weapon == Shotgun {
		t.Error("a dry shotgun should auto-switch away")
	}
	if weapons[g.Player.Weapon].ammo != ammoNone && g.ammoCount(weapons[g.Player.Weapon].ammo) == 0 {
		t.Errorf("auto-switched to an unusable weapon %v", g.Player.Weapon)
	}
}

func TestAutoSwitchFallsBackToFists(t *testing.T) {
	g := autoSwitchGame(t)
	g.Player.Weapon = Pistol
	g.Player.Bullets, g.Player.Shells, g.Player.Rockets = 0, 0, 0
	g.autoSwitchIfEmpty()
	if g.Player.Weapon != Fists {
		t.Errorf("out of all ammo should fall back to fists, got %v", g.Player.Weapon)
	}
}

func TestAutoSwitchKeepsUsableWeapon(t *testing.T) {
	g := autoSwitchGame(t)
	g.Player.Weapon = Chaingun
	g.Player.Bullets = 5
	g.autoSwitchIfEmpty()
	if g.Player.Weapon != Chaingun {
		t.Error("a weapon with ammo should not switch")
	}
}

func TestAutoSwitchNeverPicksRockets(t *testing.T) {
	g := autoSwitchGame(t)
	g.Player.Weapon = Shotgun
	g.Player.Shells, g.Player.Bullets = 0, 0
	g.Player.Rockets = 5 // only rockets left
	g.autoSwitchIfEmpty()
	if g.Player.Weapon == RocketLauncher {
		t.Error("auto-switch should not reflexively pick the rocket launcher")
	}
}
