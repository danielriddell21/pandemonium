package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestHazardFloorDrainsHealth(t *testing.T) {
	l := terrainLevel(8, 3)
	l.Hazard = map[world.Coord]world.HazardCell{{X: 4, Y: 1}: {Rate: 8.0, Kind: world.HazardNukage}}
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 4.5, Y: 1.5} // standing in the slime
	g.Player.Z = 0

	before := g.Player.Health
	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Health >= before {
		t.Errorf("hazard floor should drain health: %v -> %v", before, g.Player.Health)
	}
}

func TestHazardSafeOnLedgeAbove(t *testing.T) {
	l := terrainLevel(8, 3)
	hz := world.Coord{X: 4, Y: 1}
	l.Hazard = map[world.Coord]world.HazardCell{hz: {Rate: 8.0, Kind: world.HazardNukage}}
	setFloor(l, hz.X, hz.Y, 0) // the slime sits at the base
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 4.5, Y: 1.5}
	g.Player.Z = 1.0 // standing well above the pool (e.g. on a bridge)

	before := g.Player.Health
	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Health != before {
		t.Errorf("standing above a hazard should be safe: %v -> %v", before, g.Player.Health)
	}
}

func TestRadSuitNegatesNukageButNotLava(t *testing.T) {
	stand := func(kind world.HazardKind) (before, after float64) {
		l := terrainLevel(8, 3)
		l.Hazard = map[world.Coord]world.HazardCell{{X: 4, Y: 1}: {Rate: 8.0, Kind: kind}}
		g := New(l)
		g.Entities = nil
		g.Player.Pos = Vec2{X: 4.5, Y: 1.5}
		g.Player.Z = 0
		g.Player.RadSuitTTL = 30 // suited up
		before = g.Player.Health
		g.Tick(Input{}, 1.0/60.0)
		return before, g.Player.Health
	}
	if b, a := stand(world.HazardNukage); a != b {
		t.Errorf("a radsuit should negate nukage: %v -> %v", b, a)
	}
	if b, a := stand(world.HazardLava); a >= b {
		t.Errorf("lava should burn through a radsuit: %v -> %v", b, a)
	}
}

func TestSafeFloorDoesNotDrain(t *testing.T) {
	l := terrainLevel(8, 3)
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 4.5, Y: 1.5}
	before := g.Player.Health
	for range 60 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if g.Player.Health != before {
		t.Errorf("ordinary floor must not drain health: %v -> %v", before, g.Player.Health)
	}
}
