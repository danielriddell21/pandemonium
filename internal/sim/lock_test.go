package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func gatedGame(t *testing.T) (*Game, world.Coord, world.ItemKind) {
	t.Helper()
	for seed := int64(0); seed < 200; seed++ {
		l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		for cell, key := range l.Locks {
			return New(l), cell, key
		}
	}
	t.Skip("no gated level found in scanned seeds")
	return nil, world.Coord{}, 0
}

func faceDoor(g *Game, door world.Coord) {
	g.Player.Pos = Vec2{X: float64(door.X) + 0.5, Y: float64(door.Y) + 1.5}
	g.Player.Angle = -math.Pi / 2
}

func TestLockedDoorStaysShutWithoutKey(t *testing.T) {
	g, door, _ := gatedGame(t)
	faceDoor(g, door)
	g.Tick(Input{Interact: true}, 1.0/60.0)
	if g.World.Opened(door.X, door.Y) {
		t.Errorf("locked door %v opened without the key", door)
	}
	if g.Notice() == "" {
		t.Error("expected a 'need key' notice when interacting with a locked door")
	}
}

func TestLockedDoorOpensWithKey(t *testing.T) {
	g, door, key := gatedGame(t)
	g.Player.Keys = map[world.ItemKind]bool{key: true}
	faceDoor(g, door)
	g.Tick(Input{Interact: true}, 1.0/60.0)
	if !g.World.Opened(door.X, door.Y) {
		t.Errorf("locked door %v did not open with the %v", door, key)
	}
}
