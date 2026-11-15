package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestDiagonalMoveIsNoFasterThanStraight(t *testing.T) {
	straight := openRoom(t, 30, 30)
	straight.Player.Pos, straight.Player.Angle = Vec2{X: 15, Y: 15}, 0
	diagonal := openRoom(t, 30, 30)
	diagonal.Player.Pos, diagonal.Player.Angle = Vec2{X: 15, Y: 15}, 0

	run(straight, Input{Forward: 1}, 60)
	run(diagonal, Input{Forward: 1, Strafe: 1}, 60)

	a := dist(straight.Player.Pos, Vec2{X: 15, Y: 15})
	b := dist(diagonal.Player.Pos, Vec2{X: 15, Y: 15})
	if math.Abs(a-b) > 1e-9 {
		t.Errorf("forward+strafe covered %.4f, forward alone %.4f: diagonals must not be faster", b, a)
	}
}

func TestUnitMove(t *testing.T) {
	cases := []struct{ f, s, wantF, wantS float64 }{
		{0, 0, 0, 0},
		{0.5, 0, 0.5, 0},
		{1, 1, math.Sqrt2 / 2, math.Sqrt2 / 2},
		{-3, 0, -1, 0},
		{0, 2, 0, 1},
	}
	for _, c := range cases {
		f, s := unitMove(c.f, c.s)
		if math.Abs(f-c.wantF) > 1e-9 || math.Abs(s-c.wantS) > 1e-9 {
			t.Errorf("unitMove(%v, %v) = %v, %v; want %v, %v", c.f, c.s, f, s, c.wantF, c.wantS)
		}
	}
}

func TestBlastStopsAtWalls(t *testing.T) {
	g := openRoom(t, 20, 12)
	for y := 1; y < 11; y++ {
		wall(g, 10, y) // a full partition between the barrel and everything else
	}
	g.Player.Pos = Vec2{X: 11.5, Y: 5.5}
	g.Entities = []Entity{
		newBarrel(Vec2{X: 9.5, Y: 5.5}),
		demonAt(11.5, 6.5, Melee),
	}
	before := g.Player.Health

	g.damageEntity(0, barrelHealth)

	if g.Player.Health != before {
		t.Errorf("a blast behind a wall hurt the player: %v -> %v", before, g.Player.Health)
	}
	if g.Entities[1].Health != meleeHealth {
		t.Errorf("a blast behind a wall hurt a demon: %v", g.Entities[1].Health)
	}
}

func TestBlastFalloff(t *testing.T) {
	if got := blastFalloff(0, 2); got != 1 {
		t.Errorf("falloff at the centre = %v, want 1", got)
	}
	if got := blastFalloff(0.5, 2); got != 1 {
		t.Errorf("falloff within a quarter radius = %v, want 1", got)
	}
	if got := blastFalloff(2, 2); got != 0.25 {
		t.Errorf("falloff at the edge = %v, want 0.25", got)
	}
	if mid := blastFalloff(1.2, 2); mid <= 0.25 || mid >= 1 {
		t.Errorf("falloff mid-radius = %v, want between the edge and the centre", mid)
	}
}

func TestBlastHurtsLessWithDistance(t *testing.T) {
	hit := func(d float64) float64 {
		g := openRoom(t, 20, 12)
		g.Player.Pos = Vec2{X: 5.5, Y: 5.5}
		g.Entities = []Entity{newBarrel(Vec2{X: 5.5 + d, Y: 5.5})}
		g.damageEntity(0, barrelHealth)
		return MaxHealth - g.Player.Health
	}
	near, far := hit(0.6), hit(2.0)
	if near <= far || far <= 0 {
		t.Errorf("blast damage near=%v far=%v: want near > far > 0", near, far)
	}
}

func TestRocketBurstsOnTheNearSideOfAWall(t *testing.T) {
	g := openRoom(t, 20, 12)
	wall(g, 10, 5)
	g.Player.Pos, g.Player.Angle = Vec2{X: 4.5, Y: 5.5}, 0
	g.Player.Weapon, g.Player.Rockets = RocketLauncher, 1
	// Standing against the wall the rocket strikes: the burst must reach it.
	g.Entities = []Entity{demonAt(9.5, 6.3, Melee)}

	g.fire()
	for range 120 {
		g.advanceProjectiles(1.0 / 60.0)
	}

	if g.Entities[0].Health >= meleeHealth {
		t.Error("a rocket bursting on a wall should catch the demon standing beside it")
	}
}

func TestShotgunPutsEveryPelletIntoALoneTarget(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Pos, g.Player.Angle = Vec2{X: 4.5, Y: 5.5}, 0
	g.Player.Weapon, g.Player.Shells = Shotgun, 2
	g.Entities = []Entity{demonAt(7.5, 5.5, Baron)}

	g.fire()

	w := weapons[Shotgun]
	if got, want := baronHealth-g.Entities[0].Health, w.damage*float64(w.targets); got != want {
		t.Errorf("one shell into one demon dealt %v, want all %d pellets = %v", got, w.targets, want)
	}
	if pistol := weapons[Pistol]; w.damage*float64(w.targets) <= pistol.damage {
		t.Error("a shell must out-hit a pistol round on a single target")
	}
}

func TestDemonShotFromBeyondSightComesHunting(t *testing.T) {
	g := openRoom(t, 30, 12)
	g.Player.Pos, g.Player.Angle = Vec2{X: 3.5, Y: 5.5}, 0
	g.Entities = []Entity{demonAt(3.5+detectRadius+2, 5.5, Baron)}
	start := g.Entities[0].Pos

	run(g, Input{}, 60)
	if g.Entities[0].Pos != start {
		t.Fatal("an unprovoked demon out of sight range should hold still")
	}

	g.damageEntity(0, 1)
	run(g, Input{}, 90)
	if !g.Entities[0].Hunting() {
		t.Error("a wounded demon should be on the hunt")
	}
	if g.Entities[0].Pos.X >= start.X {
		t.Errorf("a wounded demon should close in: x %v -> %v", start.X, g.Entities[0].Pos.X)
	}
}

func TestBarrelsBlockMovement(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Pos, g.Player.Angle = Vec2{X: 4.5, Y: 5.5}, 0
	g.Entities = []Entity{newBarrel(Vec2{X: 6.5, Y: 5.5})}

	run(g, Input{Forward: 1}, 240)

	if d := dist(g.Player.Pos, g.Entities[0].Pos); d < barrelBody-1e-9 {
		t.Errorf("player walked into a barrel: %.3f from its centre", d)
	}
	if g.Player.Pos.X > 6.5 {
		t.Errorf("player passed through a barrel to x=%.2f", g.Player.Pos.X)
	}
}

func TestBodyInsideABarrelCanBackOut(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Entities = []Entity{newBarrel(Vec2{X: 6.5, Y: 5.5})}
	inside := Vec2{X: 6.6, Y: 5.5}
	if g.barrelBlocks(inside, Vec2{X: 6.8, Y: 5.5}) {
		t.Error("moving away from an overlapping barrel must be allowed")
	}
	if !g.barrelBlocks(inside, Vec2{X: 6.5, Y: 5.5}) {
		t.Error("moving deeper into a barrel must be refused")
	}
}

func TestBurstBarrelNoLongerBlocks(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Pos = Vec2{X: 15, Y: 9}
	g.Entities = []Entity{newBarrel(Vec2{X: 6.5, Y: 5.5})}
	g.damageEntity(0, barrelHealth)
	if g.barrelBlocks(Vec2{X: 5.5, Y: 5.5}, Vec2{X: 6.4, Y: 5.5}) {
		t.Error("a burst barrel should not block")
	}
}

func TestDeathRestoresTheLoadoutTheLevelWasEnteredWith(t *testing.T) {
	entry := Loadout{Health: 140, Armor: 60, Bullets: 90, Shells: 12, Rockets: 3, Backpack: true, Weapon: Chaingun}
	g := openRoom(t, 20, 12, WithLoadout(entry))
	g.Player.Bullets, g.Player.Rockets, g.Player.Armor = 0, 40, 0
	g.Player.InvulnTTL, g.Player.RadSuitTTL = 9, 9
	g.Player.Berserk = true
	g.alarm.meter, g.alarm.waves = 0.9, 3
	g.pain = 1
	g.Player.Health = 0

	g.die()

	if got := g.Loadout(); got != entry {
		t.Errorf("after death the loadout is %+v, want the entry kit %+v", got, entry)
	}
	if g.Player.Invulnerable() || g.Player.RadSuited() {
		t.Error("timed powerups must not survive a death")
	}
	if g.Alert() != 0 || g.Alarms() != 0 || g.Pain() != 0 {
		t.Errorf("alarm and pain must reset on death: alert=%v alarms=%d pain=%v", g.Alert(), g.Alarms(), g.Pain())
	}
}

func TestHeldAttackFiresAtTheWeaponsOwnPace(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Weapon, g.Player.Bullets = Chaingun, 200
	run(g, Input{Attack: true}, 60)
	spent := 200 - g.Player.Bullets
	want := int(1 / weapons[Chaingun].cooldown)
	if spent < want-1 || spent > want+1 {
		t.Errorf("holding fire for a second spent %d rounds, want about %d", spent, want)
	}
}

func TestSealedExitAndPanelPromptTargets(t *testing.T) {
	g := openRoom(t, 20, 12)
	l := g.World.Level
	panel := world.Coord{X: 6, Y: 0}
	l.Set(panel.X, panel.Y, world.TileSwitch)
	l.Switches[panel] = world.Switch{Action: world.SwitchAlarm}
	g.Player.Pos, g.Player.Angle = Vec2{X: 6.5, Y: 1.5}, -math.Pi/2
	if got := g.FacingInteractable(); got != InteractPanel {
		t.Errorf("facing the alarm panel reports %v, want InteractPanel", got)
	}
}

func TestNoDemonStartsInSightOfTheSpawn(t *testing.T) {
	for seed := int64(1); seed <= 60; seed++ {
		g := newTestGame(t, seed)
		spawn := centreOf(g.World.Level.Spawn)
		for _, e := range g.Entities {
			if e.Kind == Barrel {
				continue
			}
			if dist(e.Pos, spawn) <= detectRadius+1 && losClear(g.World, e.Pos, spawn) {
				t.Errorf("seed %d: a %d at %v can see the spawn from %.1f away", seed, e.Kind, e.Pos, dist(e.Pos, spawn))
			}
		}
		hp := g.Player.Health
		run(g, Input{}, 60)
		if g.Player.Health != hp {
			t.Errorf("seed %d: the player was hurt in the first second without moving", seed)
		}
	}
}

func TestOverlooksSpawn(t *testing.T) {
	g := openRoom(t, 40, 12)
	spawn := g.World.Level.Spawn
	if !g.overlooksSpawn(world.Coord{X: spawn.X + 4, Y: spawn.Y}) {
		t.Error("a cell close by in plain sight overlooks the spawn")
	}
	if g.overlooksSpawn(world.Coord{X: spawn.X + 20, Y: spawn.Y}) {
		t.Error("a cell far across the room does not")
	}
	for y := 1; y < 11; y++ {
		wall(g, spawn.X+2, y)
	}
	if g.overlooksSpawn(world.Coord{X: spawn.X + 4, Y: spawn.Y}) {
		t.Error("a cell behind a wall does not")
	}
}

func TestLevelWithNowhereForDemonsStillGetsItsBarrels(t *testing.T) {
	l := roomLevel(7, 7) // every floor cell is within three of the spawn
	l.Spawn = world.Coord{X: 3, Y: 3}
	l.Barrels = []world.Coord{{X: 5, Y: 5}}
	g := New(l)
	if len(g.demonFloors()) != 0 {
		t.Fatal("this room should leave no cell far enough from the spawn")
	}
	if len(g.Entities) != 1 || g.Entities[0].Kind != Barrel {
		t.Errorf("want just the barrel, got %+v", g.Entities)
	}
}
