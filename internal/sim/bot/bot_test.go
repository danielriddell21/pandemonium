package bot

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestRoamerCompletesLevels(t *testing.T) {
	// The roamer is a rough combat heuristic, not a solver: on procedurally
	// varied maps it clears most runs but snags on some when barrels and demons
	// jam a corridor. Level validity — spawn, exit and every keycard reachable —
	// is guaranteed by Generate and covered navigationally by
	// TestBFSRouteIsWalkable; this is the coarser "the bot generally gets
	// through" check, so it samples many seeds and asks for a solid majority
	// rather than a fixed count. Generation and the pilot are deterministic, so
	// the pass rate is stable.
	const seeds = 40
	const maxTicks = 6000 // 100 simulated seconds per level
	done := 0
	for seed := int64(1); seed <= seeds; seed++ {
		l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: seed})
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		g := sim.New(l)
		p := Roamer(true)
		for range maxTicks {
			g.Tick(p.Input(g), 1.0/60.0)
			if g.LevelComplete() {
				done++
				break
			}
		}
	}
	if done*2 < seeds { // fewer than half get through signals broken generation
		t.Errorf("roamer completed only %d of %d levels", done, seeds)
	}
}

func TestHunterKillsDemons(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 5})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	alive := func() int {
		n := 0
		for _, e := range g.Entities {
			if e.Alive && e.Kind != sim.Barrel {
				n++
			}
		}
		return n
	}
	start := alive()
	if start == 0 {
		t.Skip("no demons on this seed")
	}
	p := Hunter()
	for range 4000 {
		g.Tick(p.Input(g), 1.0/60.0)
		if alive() < start {
			return
		}
	}
	t.Errorf("hunter killed nothing: %d demons before and after", start)
}

func TestRoamerCompletesArena(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 40, Height: 28, Seed: 5, Arena: true})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	g.Entities = nil
	p := Roamer(false)
	for range 4000 {
		g.Tick(p.Input(g), 1.0/60.0)
		if g.LevelComplete() {
			return
		}
	}
	t.Error("roamer did not complete the arena")
}

func TestSteerTowardClosesDistance(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 24, Height: 18, Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	for _, angle := range []float64{0, math.Pi / 3, math.Pi, -math.Pi / 2} {
		g := sim.New(l)
		g.Entities = nil
		g.Player.Angle = angle
		target := sim.Vec2{X: g.Player.Pos.X + 1, Y: g.Player.Pos.Y}
		before := dist(g.Player.Pos, target)
		for range 30 {
			g.Tick(steerToward(g, target), 1.0/60.0)
		}
		if after := dist(g.Player.Pos, target); after >= before {
			t.Errorf("angle %.2f: steering widened the gap (%.2f -> %.2f)", angle, before, after)
		}
	}
}

func TestBFSRouteIsWalkable(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	route := bfsOpen(l, l.Spawn, l.Exit)
	if len(route) == 0 {
		t.Fatal("no route from spawn to exit")
	}
	if route[0] != l.Spawn || route[len(route)-1] != l.Exit {
		t.Fatalf("route endpoints %v..%v, want %v..%v", route[0], route[len(route)-1], l.Spawn, l.Exit)
	}
	for i := 1; i < len(route); i++ {
		a, b := route[i-1], route[i]
		if abs(a.X-b.X)+abs(a.Y-b.Y) != 1 {
			t.Fatalf("route step %v -> %v is not adjacent", a, b)
		}
		if !l.At(b.X, b.Y).Walkable() {
			t.Fatalf("route crosses unwalkable cell %v", b)
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
