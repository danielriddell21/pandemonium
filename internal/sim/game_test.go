package sim

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func newTestGame(t *testing.T, seed int64) *Game {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: seed})
	if err != nil {
		t.Fatal(err)
	}
	return New(l)
}

func TestPlayerStartsOnSpawn(t *testing.T) {
	g := newTestGame(t, 1)
	if got := g.PlayerCell(); got != g.World.Level.Spawn {
		t.Errorf("player cell = %v, want spawn %v", got, g.World.Level.Spawn)
	}
}

func TestPlayerNeverEntersSolidTiles(t *testing.T) {
	g := newTestGame(t, 7)
	r := rand.New(rand.NewPCG(42, 42))
	for range 5000 {
		in := Input{
			Forward: r.Float64()*2 - 1,
			Strafe:  r.Float64()*2 - 1,
			Turn:    r.Float64()*2 - 1,
		}
		g.Tick(in, 1.0/60.0)
		c := g.PlayerCell()
		if g.World.Solid(c.X, c.Y) {
			t.Fatalf("player entered solid tile %v at pos %v", c, g.Player.Pos)
		}
	}
}

func TestReachedExit(t *testing.T) {
	g := newTestGame(t, 3)
	if g.ReachedExit() {
		t.Fatal("reached exit at spawn unexpectedly")
	}
	exit := g.World.Level.Exit
	g.Player.Pos = Vec2{X: float64(exit.X) + 0.5, Y: float64(exit.Y) + 0.5}
	if !g.ReachedExit() {
		t.Error("not detecting exit when standing on it")
	}
}

func TestInteractOpensDoor(t *testing.T) {
	// Find a level that has a door, then stand in front of it and interact.
	for seed := int64(0); seed < 50; seed++ {
		l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		door, ok := firstDoor(l)
		if !ok {
			continue
		}
		g := New(l)
		// Place the player one tile away, facing the door.
		g.Player.Pos = Vec2{X: float64(door.X) + 0.5, Y: float64(door.Y) + 1.5}
		g.Player.Angle = -math.Pi / 2 // facing -Y, toward the door
		if !g.World.Solid(door.X, door.Y) {
			t.Fatalf("seed %d: door %v not solid before opening", seed, door)
		}
		g.Tick(Input{Interact: true}, 1.0/60.0)
		if g.World.Solid(door.X, door.Y) {
			t.Errorf("seed %d: door %v still solid after interact", seed, door)
		}
		return
	}
	t.Skip("no door found in scanned seeds")
}

func TestTickDeterminism(t *testing.T) {
	inputs := make([]Input, 600)
	r := rand.New(rand.NewPCG(9, 9))
	for i := range inputs {
		inputs[i] = Input{Forward: r.Float64(), Strafe: r.Float64()*2 - 1, Turn: r.Float64()*2 - 1}
	}

	run := func() (Vec2, float64) {
		g := newTestGame(t, 11)
		for _, in := range inputs {
			g.Tick(in, 1.0/60.0)
		}
		return g.Player.Pos, g.Player.Angle
	}
	p1, a1 := run()
	p2, a2 := run()
	if p1 != p2 || a1 != a2 {
		t.Errorf("non-deterministic tick: (%v,%v) vs (%v,%v)", p1, a1, p2, a2)
	}
}

func firstDoor(l *world.Level) (world.Coord, bool) {
	for y := range l.Height {
		for x := range l.Width {
			c := world.Coord{X: x, Y: y}
			if l.At(x, y) != world.TileDoor {
				continue
			}
			if _, locked := l.Locks[c]; locked {
				continue
			}
			return c, true
		}
	}
	return world.Coord{}, false
}
