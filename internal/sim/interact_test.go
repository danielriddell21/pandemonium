package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestFacingInteractable(t *testing.T) {
	g := newTestGame(t, 7)
	l := g.World.Level

	var target world.Coord
	var have bool
	for c := range l.Switches {
		target, have = c, true
		break
	}
	if !have {
		t.Skip("generated level has no switch to face")
	}

	// Stand on a non-solid neighbour of the switch and face it.
	stood := false
	for _, d := range []world.Coord{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		sx, sy := target.X-d.X, target.Y-d.Y
		if l.Solid(sx, sy) {
			continue
		}
		g.Player.Pos = Vec2{X: float64(sx) + 0.5, Y: float64(sy) + 0.5}
		g.Player.Angle = math.Atan2(float64(d.Y), float64(d.X))
		if g.FacingInteractable() == InteractSwitch {
			stood = true

			// Facing the opposite way, nothing is in reach.
			g.Player.Angle += math.Pi
			if got := g.FacingInteractable(); got == InteractSwitch {
				t.Errorf("facing away still reported a switch (got %v)", got)
			}
			break
		}
	}
	if !stood {
		t.Fatal("expected to face the switch and get InteractSwitch")
	}
}
