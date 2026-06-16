package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// terrainLevel hand-builds a small walled arena with a flat base floor, so tests
// can sculpt exact heights onto it.
func terrainLevel(w, h int) *world.Level {
	l := &world.Level{
		Width:  w,
		Height: h,
		Tiles:  make([]world.TileType, w*h),
		FloorH: make([]float64, w*h),
		CeilH:  make([]float64, w*h),
		Spawn:  world.Coord{X: 1, Y: 1},
		Exit:   world.Coord{X: w - 2, Y: h - 2},
	}
	for i := range l.CeilH {
		l.CeilH[i] = 2 // generous headroom everywhere
	}
	for y := range h {
		for x := range w {
			if x == 0 || y == 0 || x == w-1 || y == h-1 {
				l.Tiles[y*w+x] = world.TileWall
			}
		}
	}
	l.Tiles[l.Spawn.Y*w+l.Spawn.X] = world.TileSpawn
	l.Tiles[l.Exit.Y*w+l.Exit.X] = world.TileExit
	return l
}

func setFloor(l *world.Level, x, y int, f float64) {
	l.FloorH[y*l.Width+x] = f
}

func TestClimbsOneStepButNotTwo(t *testing.T) {
	l := terrainLevel(8, 3)
	setFloor(l, 4, 1, 0.25) // one step: climbable
	setFloor(l, 5, 1, 0.75) // two more steps at once: not climbable
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 2.5, Y: 1.5}
	g.Player.Angle = 0 // facing +X

	for range 300 {
		g.Tick(Input{Forward: 1}, 1.0/60.0)
	}
	if cx := int(g.Player.Pos.X); cx != 4 {
		t.Fatalf("player should stand on the climbable step (x=4), got x=%d", cx)
	}
	if g.Player.Z != 0.25 {
		t.Errorf("player Z = %v, want 0.25 after climbing the step", g.Player.Z)
	}
}

func TestFallsFromLedge(t *testing.T) {
	l := terrainLevel(8, 3)
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 3.5, Y: 1.5}
	g.Player.Z = 0.75 // dropped onto thin air above the base floor

	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Z >= 0.75 {
		t.Fatal("player did not start falling")
	}
	for range 60 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if g.Player.Z != 0 {
		t.Errorf("player Z = %v, want 0 after landing", g.Player.Z)
	}
}

func TestLiftCarriesPlayer(t *testing.T) {
	l := terrainLevel(8, 3)
	lift := world.Coord{X: 4, Y: 1}
	l.Lifts = map[world.Coord]world.Lift{lift: {Low: 0, High: 0.75}}
	setFloor(l, 5, 1, 0.75) // the ledge the lift serves
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 4.5, Y: 1.5} // standing on the platform

	// Ride through the low dwell and the upward travel.
	rise := int((liftDwell + liftTravel) / (1.0 / 60.0))
	for range rise + 10 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if math.Abs(g.Player.Z-0.75) > 1e-6 {
		t.Errorf("player Z = %v, want 0.75 riding the lift to the top", g.Player.Z)
	}
}

func TestProjectileAbsorbedByStep(t *testing.T) {
	l := terrainLevel(10, 3)
	setFloor(l, 6, 1, 0.75) // a step face taller than the fireball's flight height
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 1.5, Y: 1.5} // out of the flight path's reach

	g.Projectiles = []Projectile{{
		Pos: Vec2{X: 3.5, Y: 1.5}, Z: 0.4,
		Vel: Vec2{X: projectileSpeed}, Damage: projectileDamage, Alive: true,
	}}
	for range 60 {
		g.advanceProjectiles(1.0 / 60.0)
	}
	if len(g.Projectiles) != 0 {
		t.Error("fireball should have been absorbed by the step face")
	}
	if g.Player.Health != MaxHealth {
		t.Error("absorbed fireball must not hurt the player")
	}
}

func TestDemonCannotClimbTallLedge(t *testing.T) {
	l := terrainLevel(8, 3)
	setFloor(l, 5, 1, 0.75)
	g := New(l)
	g.Player.Pos = Vec2{X: 5.5, Y: 1.5} // player up on the ledge
	g.Player.Z = 0.75
	g.Entities = []Entity{{Pos: Vec2{X: 3.5, Y: 1.5}, Kind: Melee, State: Active, Health: meleeHealth, Alive: true}}

	for range 300 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if cx := int(g.Entities[0].Pos.X); cx >= 5 {
		t.Errorf("demon climbed the unclimbable ledge: x=%d", cx)
	}
	if g.Player.Health != MaxHealth {
		t.Errorf("demon below a ledge should not claw the player: health %v", g.Player.Health)
	}
}

func TestEyeFollowsTerrain(t *testing.T) {
	l := terrainLevel(8, 3)
	setFloor(l, 3, 1, 0.25)
	g := New(l)
	g.Entities = nil
	if g.EyeZ() != eyeHeight {
		t.Fatalf("EyeZ at spawn = %v, want %v", g.EyeZ(), eyeHeight)
	}
	g.Player.Pos = Vec2{X: 3.5, Y: 1.5}
	for range 120 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if math.Abs(g.EyeZ()-(0.25+eyeHeight)) > 1e-6 {
		t.Errorf("EyeZ = %v, want %v after settling on the step", g.EyeZ(), 0.25+eyeHeight)
	}
}
