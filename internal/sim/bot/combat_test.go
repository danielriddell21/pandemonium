package bot

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestThreatPicksTheNearestVisibleDemon(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos = sim.Vec2{X: 5.5, Y: 5.5}
	if _, ok := threat(g); ok {
		t.Fatal("an empty room holds no threat")
	}
	g.Entities = []sim.Entity{
		demon(11.5, 5.5, sim.Melee),
		demon(8.5, 5.5, sim.Ranged),
		{Pos: sim.Vec2{X: 6.5, Y: 5.5}, Kind: sim.Barrel, State: sim.Active, Alive: true},
		{Pos: sim.Vec2{X: 6.0, Y: 5.5}, Kind: sim.Pinky},
		demon(5.5+sightRange+1, 5.5, sim.Gunner),
	}
	e, ok := threat(g)
	if !ok || e.Kind != sim.Ranged {
		t.Fatalf("threat = %+v, want the nearer imp; barrels, corpses and far demons do not count", e)
	}
	for y := 1; y < 11; y++ {
		g.World.Level.Set(7, y, world.TileWall)
	}
	if _, ok := threat(g); ok {
		t.Error("a demon behind a wall is no threat yet")
	}
}

func TestBossIsAThreatFromAcrossTheRoom(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos = sim.Vec2{X: 5.5, Y: 5.5}
	g.Entities = []sim.Entity{demon(5.5+sightRange+2, 5.5, sim.Boss)}
	if _, ok := threat(g); !ok {
		t.Error("the overlord should be engaged from beyond ordinary sight range")
	}
}

func TestPickWeapon(t *testing.T) {
	imp, baron := demon(0, 0, sim.Ranged), demon(0, 0, sim.Baron)
	armed := sim.Player{Bullets: 50, Shells: 10, Rockets: 5}
	cases := []struct {
		name string
		p    sim.Player
		e    sim.Entity
		d    float64
		lane bool
		want sim.WeaponKind
	}{
		{"rockets at range with a clear lane", armed, imp, 6, true, sim.RocketLauncher},
		{"no rockets down a blocked lane", armed, imp, 6, false, sim.Shotgun},
		{"no rockets up close", armed, imp, 3.5, true, sim.Shotgun},
		{"chaingun for a baron beyond shotgun range", sim.Player{Bullets: 50, Shells: 10}, baron, 8, true, sim.Chaingun},
		{"chaingun in a pinch", sim.Player{Bullets: 50}, imp, 2, true, sim.Chaingun},
		{"pistol to save bullets", sim.Player{Bullets: 50}, imp, 6, true, sim.Pistol},
		{"shotgun even out of range when it is all there is", sim.Player{Shells: 3}, imp, 8.5, true, sim.Shotgun},
		{"fists when dry", sim.Player{}, imp, 4, true, sim.Fists},
		{"berserk fists up close", sim.Player{Berserk: true, Bullets: 50, Shells: 10}, imp, 1, true, sim.Fists},
	}
	for _, c := range cases {
		if got := pickWeapon(c.p, c.e, c.d, c.lane); got != c.want {
			t.Errorf("%s: picked %v, want %v", c.name, got, c.want)
		}
	}
}

func TestGunnerShootsWhatItFaces(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 5.5}, 0
	g.Player.Weapon = sim.Shotgun
	g.Entities = []sim.Entity{demon(10.5, 5.5, sim.Melee)}
	var c gunner
	in, ok := c.fight(g)
	if !ok || !in.Attack {
		t.Fatalf("facing a demon in range the pilot should fire: %+v ok=%v", in, ok)
	}
	if in.Strafe == 0 {
		t.Error("the pilot should keep moving across the line of fire")
	}
	if in.Forward != 0 {
		t.Errorf("at a comfortable range the pilot should hold its ground: forward %v", in.Forward)
	}
}

func TestGunnerTurnsBeforeItShoots(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 5.5}, math.Pi
	g.Player.Weapon = sim.Shotgun
	g.Entities = []sim.Entity{demon(10.5, 5.5, sim.Melee)}
	var c gunner
	in, _ := c.fight(g)
	if in.Attack || in.TurnDelta == 0 {
		t.Errorf("with its back to the demon the pilot should turn, not fire: %+v", in)
	}
	if math.Abs(in.TurnDelta) > maxTurn {
		t.Errorf("turn %v exceeds the cap %v", in.TurnDelta, maxTurn)
	}
}

func TestGunnerSwitchesWeaponThenFires(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 5.5}, 0
	g.Player.Weapon, g.Player.Shells = sim.Pistol, 10
	g.Entities = []sim.Entity{demon(10.5, 5.5, sim.Melee)}
	var c gunner
	in, _ := c.fight(g)
	if in.SelectWeapon != int(sim.Shotgun)+1 || in.Attack {
		t.Errorf("the pilot should draw the shotgun before firing: %+v", in)
	}
}

func TestGunnerBacksOffAndClosesIn(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 5.5}, 0
	g.Player.Weapon = sim.Shotgun
	g.Entities = []sim.Entity{demon(7, 5.5, sim.Melee)}
	var c gunner
	if in, _ := c.fight(g); in.Forward != -1 {
		t.Errorf("crowded, the pilot should back off: %+v", in)
	}
	// Boxed in against a wall there is nowhere to back off to.
	g.Player.Pos = sim.Vec2{X: 1.3, Y: 5.5}
	g.Entities = []sim.Entity{demon(2.8, 5.5, sim.Melee)}
	if in, _ := c.fight(g); in.Forward != 0 {
		t.Errorf("against a wall the pilot cannot retreat: %+v", in)
	}
	// Out of shells and bullets: only fists, so it must close the gap.
	g.Player.Pos = sim.Vec2{X: 5.5, Y: 5.5}
	g.Player.Bullets, g.Player.Shells, g.Player.Weapon = 0, 0, sim.Fists
	g.Entities = []sim.Entity{demon(10.5, 5.5, sim.Melee)}
	if in, _ := c.fight(g); in.Forward != 1 || in.Attack {
		t.Errorf("out of reach with fists the pilot should advance: %+v", in)
	}
}

func TestGunnerWillNotBurstABarrelInItsOwnFace(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 5.5}, 0
	g.Player.Weapon = sim.Shotgun
	g.Entities = []sim.Entity{
		demon(10.5, 5.5, sim.Melee),
		{Pos: sim.Vec2{X: 7, Y: 5.5}, Kind: sim.Barrel, State: sim.Active, Alive: true},
	}
	var c gunner
	in, ok := c.fight(g)
	if !ok || in.Attack || in.Forward != -1 {
		t.Fatalf("with a barrel at its feet the pilot should step back and hold fire: %+v ok=%v", in, ok)
	}
	// No room to step back: leave the fight for a while rather than dither.
	g.Player.Pos = sim.Vec2{X: 1.3, Y: 5.5}
	g.Entities[0].Pos = sim.Vec2{X: 6.5, Y: 5.5}
	g.Entities[1].Pos = sim.Vec2{X: 2.8, Y: 5.5}
	if _, ok := c.fight(g); ok {
		t.Fatal("boxed in behind a barrel the pilot should walk on")
	}
	if _, ok := c.fight(g); ok || c.ignore != ignoreTicks-1 {
		t.Errorf("the pilot should stay out of the fight for a spell: ignore=%d", c.ignore)
	}
}

func TestBarrelInLine(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 5.5}, 0
	barrel := sim.Entity{Kind: sim.Barrel, State: sim.Active, Alive: true}
	at := func(x, y float64) bool {
		barrel.Pos = sim.Vec2{X: x, Y: y}
		g.Entities = []sim.Entity{barrel}
		return barrelInLine(g, 7, 0.82)
	}
	if !at(7, 5.5) {
		t.Error("a barrel dead ahead and close is in the line of fire")
	}
	if at(5.5+barrelDanger+0.5, 5.5) {
		t.Error("a barrel beyond its blast reach is fair game")
	}
	if at(5.5, 7) {
		t.Error("a barrel off to the side is not in the line of fire")
	}
	if at(5.5, 5.5) {
		t.Error("a barrel at zero range has no direction and is skipped")
	}
}

func TestRocketLane(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos = sim.Vec2{X: 5.5, Y: 5.5}
	to := sim.Vec2{X: 8, Y: 0}
	if !rocketLane(g, to, 8) {
		t.Fatal("an empty room is a clear lane")
	}
	g.Entities = []sim.Entity{demon(7.5, 5.9, sim.Melee)}
	if rocketLane(g, to, 8) {
		t.Error("a demon standing beside the lane would set the rocket off")
	}
	g.Entities = []sim.Entity{demon(13.5, 5.5, sim.Melee)}
	if !rocketLane(g, to, 8) {
		t.Error("the target itself, beyond the safe distance, does not block the lane")
	}
	g.Entities = nil
	g.World.Level.Set(7, 5, world.TileWall)
	if rocketLane(g, to, 8) {
		t.Error("a wall inside the safe distance blocks the lane")
	}
	g.World.Level.Set(7, 5, world.TileFloor)
	g.World.Level.SetFloor(7, 5, 0.75)
	if rocketLane(g, to, 8) {
		t.Error("a step up to eye height blocks the lane")
	}
}

func TestSidestepKeepsTheTargetInSightAndOffThePools(t *testing.T) {
	g := room(t, 30, 12)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 5.5}, 0
	target := sim.Vec2{X: 12.5, Y: 5.5}
	var c gunner
	first := c.sidestep(g, target)
	if first == 0 {
		t.Fatal("in the open the pilot should sidestep")
	}
	// Pools on both flanks: nowhere to slide.
	l := g.World.Level
	l.Hazard[world.Coord{X: 5, Y: 4}] = world.HazardCell{Rate: 8}
	l.Hazard[world.Coord{X: 5, Y: 6}] = world.HazardCell{Rate: 8}
	if got := c.sidestep(g, target); got != 0 {
		t.Errorf("with pools either side the pilot should stand, got %v", got)
	}
	if c.clear(g, 0, 1) || c.clear(g, 0, -1) {
		t.Error("a pool is not clear ground")
	}
	if !c.clear(g, -1, 0) {
		t.Error("the open floor behind is clear")
	}
	if p := probe(g, 1, 0); math.Abs(p.X-(5.5+sideProbe)) > 1e-9 || math.Abs(p.Y-5.5) > 1e-9 {
		t.Errorf("probe ahead = %v", p)
	}
}

func TestSidestepTurnsAboutOnATimer(t *testing.T) {
	g := room(t, 30, 30)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 15.5, Y: 15.5}, 0
	target := sim.Vec2{X: 22.5, Y: 15.5}
	var c gunner
	seen := map[float64]bool{}
	for range 3 * strafeTicks {
		g.Player.Pos.Y += 0.02 // keep moving so only the timer flips the side
		seen[c.sidestep(g, target)] = true
	}
	if !seen[1] || !seen[-1] {
		t.Errorf("the sidestep should swing both ways over time: %v", seen)
	}
}
