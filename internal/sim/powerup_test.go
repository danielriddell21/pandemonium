package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestSoulsphereOverheals(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Health = 80
	g.applyPickup(world.ItemSoul)
	if g.Player.Health != 180 {
		t.Errorf("soulsphere should overheal to 180, got %v", g.Player.Health)
	}
}

func TestMegasphereMaxesHealthAndArmour(t *testing.T) {
	g := newTestGame(t, 7)
	g.applyPickup(world.ItemMega)
	if g.Player.Health != overHealMax || g.Player.Armor != overHealMax {
		t.Errorf("megasphere should max both: hp=%v armor=%v", g.Player.Health, g.Player.Armor)
	}
}

func TestInvulnerabilityIgnoresDamageThenExpires(t *testing.T) {
	g := newTestGame(t, 7)
	g.applyPickup(world.ItemInvuln)
	g.Player.Health = 100
	g.hurtPlayer(50)
	if g.Player.Health != 100 {
		t.Errorf("invulnerable player should take no damage, got %v", g.Player.Health)
	}
	g.Player.InvulnTTL = 0 // expire
	g.hurtPlayer(50)
	if g.Player.Health == 100 {
		t.Error("damage should land once invulnerability has expired")
	}
}

func TestRadSuitNegatesHazardFloor(t *testing.T) {
	l := terrainLevel(8, 3)
	l.Hazard = map[world.Coord]float64{{X: 4, Y: 1}: 8.0}
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 4.5, Y: 1.5}
	g.Player.RadSuitTTL = 10
	before := g.Player.Health
	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Health != before {
		t.Errorf("radiation suit should block hazard damage: %v -> %v", before, g.Player.Health)
	}
}

func TestBerserkBoostsFists(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Weapon = Fists
	g.Player.Angle = 0
	front := Vec2{X: g.Player.Pos.X + 1.4, Y: g.Player.Pos.Y}
	if g.World.Solid(int(front.X), int(front.Y)) || !losClear(g.World, g.Player.Pos, front) {
		t.Skip("no clear space ahead")
	}
	g.Entities = []Entity{{Pos: front, Kind: Melee, State: Active, Health: 150, Alive: true}}
	g.applyPickup(world.ItemBerserk)
	g.fire() // one berserk punch should fell a 150-HP demon (200 dmg)
	if g.Entities[0].State == Active {
		t.Error("a berserk punch should one-shot the demon")
	}
}

func TestPowerupTimersCountDown(t *testing.T) {
	g := newTestGame(t, 7)
	g.Entities = nil
	g.Player.InvulnTTL = 1.0
	g.Player.RadSuitTTL = 1.0
	g.tickPowerups(0.5)
	if g.Player.InvulnTTL > 0.6 || g.Player.RadSuitTTL > 0.6 {
		t.Errorf("powerup timers should decrement: invuln=%v rad=%v", g.Player.InvulnTTL, g.Player.RadSuitTTL)
	}
}
