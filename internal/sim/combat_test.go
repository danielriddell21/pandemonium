package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestContactDamageDrainsHealth(t *testing.T) {
	g := newTestGame(t, 7)
	// Put a demon right on top of the player.
	g.Entities = []Entity{{Pos: g.Player.Pos, Alive: true}}
	before := g.Player.Health
	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Health >= before {
		t.Errorf("health did not drop on contact: %v -> %v", before, g.Player.Health)
	}
}

func TestNoDamageWithoutDemons(t *testing.T) {
	g := newTestGame(t, 7)
	g.Entities = nil
	for range 120 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if g.Player.Health != MaxHealth {
		t.Errorf("health changed with no demons: %v", g.Player.Health)
	}
}

func TestDemonChasesTowardPlayer(t *testing.T) {
	g := newTestGame(t, 7)
	p := g.Player.Pos
	// Place a demon a few tiles away with clear line of sight, then verify it
	// closes the distance over a second of ticks.
	var start Vec2
	found := false
	for r := 2.0; r <= 5; r += 0.5 {
		cand := Vec2{X: p.X + r, Y: p.Y}
		if !g.World.Solid(int(cand.X), int(cand.Y)) && losClear(g.World, cand, p) {
			start = cand
			found = true
			break
		}
	}
	if !found {
		t.Skip("no clear demon placement for this level")
	}
	g.Entities = []Entity{{Pos: start, Alive: true}}
	before := dist(start, p)
	for range 60 {
		g.Tick(Input{}, 1.0/60.0)
	}
	after := dist(g.Entities[0].Pos, g.Player.Pos)
	if after >= before {
		t.Errorf("demon did not approach: %v -> %v", before, after)
	}
}

func TestDemonBlockedByWall(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 5})
	if err != nil {
		t.Fatal(err)
	}
	g := New(l)
	// A demon outside the world bounds (surrounded by solid) must not move.
	g.Entities = []Entity{{Pos: Vec2{X: 0.5, Y: 0.5}, Alive: true}} // corner, walls around
	start := g.Entities[0].Pos
	for range 30 {
		g.Tick(Input{}, 1.0/60.0)
	}
	if g.Entities[0].Pos != start {
		t.Errorf("walled-in demon moved: %v -> %v", start, g.Entities[0].Pos)
	}
}
