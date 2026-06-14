package sim

import (
	"math"
	"testing"
)

func TestPistolConsumesBullets(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Weapon = Pistol
	g.Player.Bullets = 3
	if !g.fire() {
		t.Fatal("pistol with ammo should fire")
	}
	if g.Player.Bullets != 2 {
		t.Errorf("bullets = %d, want 2", g.Player.Bullets)
	}
	if !g.MuzzleFlash() {
		t.Error("firing should raise a muzzle flash")
	}
}

func TestEmptyGunDoesNotFire(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Weapon = Shotgun
	g.Player.Shells = 0
	if g.fire() {
		t.Error("a shotgun with no shells should not fire")
	}
	if g.MuzzleFlash() {
		t.Error("no flash should appear when out of ammo")
	}
}

func TestFistsNeedNoAmmo(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Weapon = Fists
	g.Player.Bullets, g.Player.Shells = 0, 0
	if !g.fire() {
		t.Error("fists should always swing")
	}
}

func TestShotgunHitsAClusteredGroup(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Weapon = Shotgun
	g.Player.Shells = 5
	p := g.Player.Pos

	ok := false
	for _, ang := range []float64{0, math.Pi / 2, math.Pi, -math.Pi / 2, math.Pi / 4, -math.Pi / 4} {
		ux, uy := math.Cos(ang), math.Sin(ang)
		a := Vec2{X: p.X + 2*ux, Y: p.Y + 2*uy}
		b := Vec2{X: p.X + 3*ux, Y: p.Y + 3*uy}
		c := Vec2{X: p.X + 4*ux, Y: p.Y + 4*uy}
		if g.World.Solid(int(c.X), int(c.Y)) || !losClear(g.World, p, c) {
			continue
		}
		g.Player.Angle = ang
		g.Entities = []Entity{
			{Pos: a, State: Active, Health: meleeHealth, Alive: true},
			{Pos: b, State: Active, Health: meleeHealth, Alive: true},
			{Pos: c, State: Active, Health: meleeHealth, Alive: true},
		}
		ok = true
		break
	}
	if !ok {
		t.Skip("no clear line for the shotgun spread test")
	}

	g.fire()
	wounded := 0
	for _, e := range g.Entities {
		if e.Health < meleeHealth {
			wounded++
		}
	}
	if wounded < 2 {
		t.Errorf("shotgun should spread across the group; wounded %d of 3", wounded)
	}
}
