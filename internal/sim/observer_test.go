package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

type captureObserver struct {
	events []Observation
}

func (c *captureObserver) Observe(o Observation) { c.events = append(c.events, o) }

func (c *captureObserver) count(k ObservationKind) int {
	n := 0
	for _, e := range c.events {
		if e.Kind == k {
			n++
		}
	}
	return n
}

func TestObserverDefaultsToNop(t *testing.T) {
	// A game with no observer option must not panic when it emits.
	g := newTestGame(t, 5)
	if _, ok := g.observer.(nopObserver); !ok {
		t.Fatalf("default observer = %T, want nopObserver", g.observer)
	}
	g.Tick(Input{Forward: 1}, 1.0/60.0)
}

func TestObserverFiresOnMovement(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	capt := &captureObserver{}
	g := New(l, WithObserver(capt))

	// Drive the player forward enough to cross at least one tile boundary.
	for range 240 {
		g.Tick(Input{Forward: 1}, 1.0/60.0)
	}
	if capt.count(ObsMove) == 0 {
		t.Fatal("expected at least one move observation")
	}
	for _, e := range capt.events {
		if e.Tick == 0 {
			t.Error("observation missing tick stamp")
		}
	}
}

func TestObserverFiresOnExit(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 40, Height: 28, Seed: 2})
	if err != nil {
		t.Fatal(err)
	}
	capt := &captureObserver{}
	g := New(l, WithObserver(capt))

	// Stand on the exit tile and throw the exit switch (or, on the rare level with
	// no switch, simply stand on the exit pad).
	exit := l.Exit
	g.Player.Pos = Vec2{X: float64(exit.X) + 0.5, Y: float64(exit.Y) + 0.5}
	in := Input{}
	for c, s := range l.Switches {
		if s.Action == world.SwitchExit {
			g.Player.Angle = math.Atan2(float64(c.Y-exit.Y), float64(c.X-exit.X))
			in.Interact = true
		}
	}
	g.Tick(in, 1.0/60.0)

	if capt.count(ObsExit) != 1 {
		t.Errorf("exit observations = %d, want 1", capt.count(ObsExit))
	}
	// A second tick must not re-fire it.
	g.Tick(Input{}, 1.0/60.0)
	if capt.count(ObsExit) != 1 {
		t.Errorf("exit re-fired: count = %d, want 1", capt.count(ObsExit))
	}
}

func TestObserverFiresOnMarkerCrossing(t *testing.T) {
	// Find a level with a junction marker, place the player on it, and confirm
	// the marker observation fires with branch information.
	for seed := int64(0); seed < 300; seed++ {
		l, err := world.Generate(world.Config{Width: 64, Height: 48, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		mk, ok := firstMarker(l, world.MarkerJunction)
		if !ok {
			continue
		}
		capt := &captureObserver{}
		g := New(l, WithObserver(capt))
		// Stand just off the junction tile, then step onto it.
		g.Player.Pos = Vec2{X: float64(mk.At.X) + 0.5, Y: float64(mk.At.Y) + 0.5}
		g.tracker.lastCell = world.Coord{X: mk.At.X, Y: mk.At.Y + 1}
		g.tracker.started = true
		g.observeMovement()

		found := false
		for _, e := range capt.events {
			if e.Kind == ObsMarker && e.Marker == world.MarkerJunction && e.At == mk.At {
				found = true
				if len(e.Ignored)+1 != len(mk.Branches) {
					t.Errorf("taken+ignored = %d, want %d branches", len(e.Ignored)+1, len(mk.Branches))
				}
			}
		}
		if !found {
			t.Errorf("seed %d: junction marker observation not emitted", seed)
		}
		return
	}
	t.Skip("no junction marker found in scanned seeds")
}

func firstMarker(l *world.Level, kind world.MarkerKind) (world.Marker, bool) {
	for _, m := range l.Markers {
		if m.Kind == kind {
			return m, true
		}
	}
	return world.Marker{}, false
}
