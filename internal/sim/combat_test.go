package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestDeathFiresObsDeathAndRespawns(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 40, Height: 28, Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	cap := &captureObserver{}
	g := New(l, WithObserver(cap))

	// A demon glued to the player drains health to zero within a few seconds.
	g.Entities = []Entity{{Pos: g.Player.Pos, Alive: true}}
	for i := 0; i < 1200 && cap.count(ObsDeath) == 0; i++ {
		g.Tick(Input{}, 1.0/60.0)
	}

	if cap.count(ObsDeath) != 1 {
		t.Fatalf("ObsDeath count = %d, want 1", cap.count(ObsDeath))
	}
	if g.Player.Health != MaxHealth {
		t.Errorf("not respawned at full health: %v", g.Player.Health)
	}
	if g.PlayerCell() != l.Spawn {
		t.Errorf("not respawned at spawn: %v vs %v", g.PlayerCell(), l.Spawn)
	}
}

func TestAttackKillsNearestDemonInFront(t *testing.T) {
	g := newTestGame(t, 7)
	p := g.Player.Pos

	// Find a facing where 1.5 and 2.7 tiles ahead are open with line of sight.
	ok := false
	for _, ang := range []float64{0, math.Pi / 2, math.Pi, -math.Pi / 2, math.Pi / 4, -math.Pi / 4} {
		a := Vec2{X: p.X + 1.5*math.Cos(ang), Y: p.Y + 1.5*math.Sin(ang)}
		b := Vec2{X: p.X + 2.7*math.Cos(ang), Y: p.Y + 2.7*math.Sin(ang)}
		if g.World.Solid(int(a.X), int(a.Y)) || g.World.Solid(int(b.X), int(b.Y)) {
			continue
		}
		if !losClear(g.World, p, a) || !losClear(g.World, p, b) {
			continue
		}
		g.Player.Angle = ang
		g.Entities = []Entity{{Pos: a, Alive: true}, {Pos: b, Alive: true}}
		ok = true
		break
	}
	if !ok {
		t.Skip("no clear line for attack test on this level")
	}

	g.Tick(Input{Attack: true}, 1.0/60.0)
	if g.Entities[0].Alive {
		t.Error("nearest demon in front should be dead")
	}
	if !g.Entities[1].Alive {
		t.Error("only the nearest demon should be hit, not the one behind it")
	}
}

func TestAttackMissesDemonBehind(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Angle = 0 // facing +X
	p := g.Player.Pos
	behind := Vec2{X: p.X - 1.5, Y: p.Y}
	if g.World.Solid(int(behind.X), int(behind.Y)) || !losClear(g.World, p, behind) {
		t.Skip("no clear space behind player")
	}
	g.Entities = []Entity{{Pos: behind, Alive: true}}
	g.Tick(Input{Attack: true}, 1.0/60.0)
	if !g.Entities[0].Alive {
		t.Error("demon behind the player should not be hit")
	}
}

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
