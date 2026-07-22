package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func gatedExitGame(t *testing.T) (*Game, float64) {
	t.Helper()
	for seed := int64(0); seed < 60; seed++ {
		l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if !l.HasExitSwitch() {
			continue
		}
		g := New(l)
		g.Player.Pos = Vec2{X: float64(l.Exit.X) + 0.5, Y: float64(l.Exit.Y) + 0.5}
		for c, s := range l.Switches {
			if s.Action == world.SwitchExit {
				return g, math.Atan2(float64(c.Y-l.Exit.Y), float64(c.X-l.Exit.X))
			}
		}
	}
	t.Skip("no level with an exit switch found")
	return nil, 0
}

func TestStandingOnExitIsNotEnough(t *testing.T) {
	g, _ := gatedExitGame(t)
	g.Player.Angle = math.Pi // looking away from the switch
	for range 30 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if g.LevelComplete() {
		t.Error("standing on the exit without pressing the switch should not finish the level")
	}
}

func TestExitSwitchFinishesLevel(t *testing.T) {
	g, face := gatedExitGame(t)
	g.Player.Angle = face
	g.Tick(Input{Interact: true}, 1.0/60.0)
	if !g.LevelComplete() {
		t.Error("pressing the exit switch should finish the level")
	}
}

func TestRemoteSwitchOpensDoor(t *testing.T) {
	// A hand-built level: a switch wall whose target is a closed door elsewhere.
	l := terrainLevel(8, 3)
	door := world.Coord{X: 5, Y: 1}
	sw := world.Coord{X: 3, Y: 0} // a wall cell on the top border
	l.Tiles[door.Y*l.W+door.X] = world.TileDoor
	l.Tiles[sw.Y*l.W+sw.X] = world.TileSwitch
	l.Switches = map[world.Coord]world.Switch{sw: {Action: world.SwitchDoor, Target: door}}
	g := New(l)
	g.Entities = nil

	if !g.World.Solid(door.X, door.Y) {
		t.Fatal("door should start closed/solid")
	}
	// Stand under the switch and press up into it.
	g.Player.Pos = Vec2{X: 3.5, Y: 1.5}
	g.Player.Angle = -math.Pi / 2
	g.Tick(Input{Interact: true}, 1.0/60.0)
	if g.World.Solid(door.X, door.Y) {
		t.Error("remote switch should have opened its target door")
	}
}
