package sim

import (
	"testing"

	"github.com/danielriddell21/crucible/level"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func openRoom(t *testing.T, w, h int, opts ...Option) *Game {
	t.Helper()
	g := New(roomLevel(w, h), opts...)
	g.Entities = nil
	g.Items = nil
	return g
}

func bossRoom(t *testing.T, opts ...Option) (*Game, *Entity) {
	t.Helper()
	l := roomLevel(40, 20)
	l.Boss, l.BossAt = true, world.Coord{X: 30, Y: 10}
	g := New(l, opts...)
	g.Items = nil
	// Only the overlord: the escort would muddy what each test measures.
	for _, e := range g.Entities {
		if e.Kind == Boss {
			g.Entities = []Entity{e}
			return g, &g.Entities[0]
		}
	}
	t.Fatal("boss floor spawned no boss")
	return nil, nil
}

func roomLevel(w, h int) *world.Level {
	base := level.New(w, h, 1)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			base.Set(x, y, world.TileFloor)
		}
	}
	base.Spawn = world.Coord{X: 2, Y: h / 2}
	base.Exit = world.Coord{X: w - 3, Y: h / 2}
	return &world.Level{
		Level:    base,
		Locks:    map[world.Coord]world.ItemKind{},
		Hazard:   map[world.Coord]world.HazardCell{},
		Switches: map[world.Coord]world.Switch{},
	}
}

func demonAt(x, y float64, kind EntityKind) Entity {
	return newDemon(Vec2{X: x, Y: y}, kind)
}

func wall(g *Game, x, y int) {
	g.World.Level.Set(x, y, world.TileWall)
}

func run(g *Game, in Input, ticks int) {
	for range ticks {
		g.Tick(in, 1.0/60.0)
	}
}

type recorder struct {
	kinds []ObservationKind
}

func (r *recorder) Observe(o Observation) { r.kinds = append(r.kinds, o.Kind) }

func (r *recorder) count(k ObservationKind) int {
	n := 0
	for _, got := range r.kinds {
		if got == k {
			n++
		}
	}
	return n
}
