package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestHazardFloorDrainsHealth(t *testing.T) {
	l := terrainLevel(8, 3)
	l.Hazard = map[world.Coord]float64{{X: 4, Y: 1}: 8.0}
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
	l.Hazard = map[world.Coord]float64{hz: 8.0}
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
