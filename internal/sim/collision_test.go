package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// TestCanDescendIntoLowerCeilingCell guards a movement bug: stepping down from a
// raised floor into a lower cell whose ceiling clears its own floor must be
// allowed, even though the ceiling is closer than the headroom margin to the
// mover's (higher) starting height.
func TestCanDescendIntoLowerCeilingCell(t *testing.T) {
	l := terrainLevel(8, 3)
	// A ledge at 0.5 stepping down to a corridor-height cell at 0.25.
	setFloor(l, 4, 1, 0.5)
	setFloor(l, 5, 1, 0.25)
	l.CeilH[1*l.Width+5] = 1.15 // ~corridor headroom above the 0.25 floor (0.9)

	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 4.5, Y: 1.5}
	g.Player.Z = 0.5
	g.Player.Angle = 0 // +X, toward the lower cell

	for range 40 {
		g.Tick(Input{Forward: 1}, 1.0/60.0)
	}
	if g.PlayerCell().X < 5 {
		t.Fatalf("player should have descended into the lower cell, stuck at %v", g.PlayerCell())
	}
}

// TestCannotEnterTrulyCrampedCell confirms the headroom rule still blocks cells
// that are too short to stand in on their own terms.
func TestCannotEnterTrulyCrampedCell(t *testing.T) {
	l := terrainLevel(8, 3)
	l.CeilH[1*l.Width+5] = world.MinHeadroom - 0.2 // ceiling below standing height
	g := New(l)
	g.Entities = nil
	g.Player.Pos = Vec2{X: 4.5, Y: 1.5}
	g.Player.Angle = 0
	for range 40 {
		g.Tick(Input{Forward: 1}, 1.0/60.0)
	}
	if g.PlayerCell().X >= 5 {
		t.Errorf("player squeezed into a cell too cramped to stand in: %v", g.PlayerCell())
	}
}
