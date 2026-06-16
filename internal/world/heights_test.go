package world

import (
	"math"
	"testing"
)

// ledgeCells returns the lift-served ledge cells of a level (the high side of
// each lift), which are exempt from the single-step invariant.
func ledgeCells(l *Level) map[Coord]bool {
	out := make(map[Coord]bool)
	for neck, lift := range l.Lifts {
		for _, n := range neighbors4(neck) {
			if l.At(n.X, n.Y).Walkable() && math.Abs(l.Floor(n.X, n.Y)-lift.High) < 1e-9 {
				out[n] = true
			}
		}
	}
	return out
}

func TestHeightsKeepSingleStepsAndHeadroom(t *testing.T) {
	for seed := int64(0); seed < 200; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		ledges := ledgeCells(l)
		for y := range l.Height {
			for x := range l.Width {
				c := Coord{X: x, Y: y}
				if !l.At(x, y).Walkable() {
					continue
				}
				if l.Ceil(x, y)-l.Floor(x, y) < MinHeadroom-1e-9 {
					t.Fatalf("seed %d: cell %v lacks headroom: floor %.2f ceil %.2f",
						seed, c, l.Floor(x, y), l.Ceil(x, y))
				}
				if ledges[c] {
					continue // reached by lift, not by steps
				}
				for _, n := range neighbors4(c) {
					if !l.At(n.X, n.Y).Walkable() || ledges[n] {
						continue
					}
					d := math.Abs(l.Floor(x, y) - l.Floor(n.X, n.Y))
					if d > StepHeight+1e-9 {
						t.Fatalf("seed %d: floors %v->%v jump %.2f (> one step)", seed, c, n, d)
					}
				}
			}
		}
	}
}

func TestExitStandsOnDais(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		e := l.Exit
		ef := l.Floor(e.X, e.Y)
		// The exit must sit above at least one walkable neighbour, with that
		// neighbour exactly one step below (the dais ring).
		raised := false
		for _, n := range neighbors4(e) {
			if !l.At(n.X, n.Y).Walkable() {
				continue
			}
			d := ef - l.Floor(n.X, n.Y)
			if d > 1e-9 {
				raised = true
				if d > StepHeight+1e-9 {
					t.Fatalf("seed %d: dais ring %v is %.2f below the exit (> one step)", seed, n, d)
				}
			}
		}
		if !raised {
			t.Errorf("seed %d: exit %v is not raised above any neighbour", seed, e)
		}
	}
}

func TestHeightsAreDeterministic(t *testing.T) {
	cfg := Config{Width: 48, Height: 32, Seed: 17}
	a, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for i := range a.FloorH {
		if a.FloorH[i] != b.FloorH[i] || a.CeilH[i] != b.CeilH[i] {
			t.Fatalf("heights differ at index %d between identical seeds", i)
		}
	}
	if len(a.Lifts) != len(b.Lifts) {
		t.Fatalf("lift count differs: %d vs %d", len(a.Lifts), len(b.Lifts))
	}
}

func TestLiftServesItsLedge(t *testing.T) {
	found := false
	for seed := int64(0); seed < 200 && !found; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		for neck, lift := range l.Lifts {
			found = true
			if lift.High-lift.Low <= MaxStep {
				t.Errorf("seed %d: lift at %v rises %.2f, climbable without it", seed, neck, lift.High-lift.Low)
			}
			if math.Abs(l.Floor(neck.X, neck.Y)-lift.Low) > 1e-9 {
				t.Errorf("seed %d: lift tile floor %.2f != Low %.2f", seed, l.Floor(neck.X, neck.Y), lift.Low)
			}
			// The shaft must clear a body riding at the top.
			if l.Ceil(neck.X, neck.Y)-lift.High < MinHeadroom-1e-9 {
				t.Errorf("seed %d: lift shaft at %v lacks headroom at the top", seed, neck)
			}
			// The ledge it serves holds a reward.
			ledge := false
			for _, n := range neighbors4(neck) {
				if math.Abs(l.Floor(n.X, n.Y)-lift.High) < 1e-9 && itemAt(l, n) {
					ledge = true
				}
			}
			if !ledge {
				t.Errorf("seed %d: lift at %v serves no rewarded ledge", seed, neck)
			}
			// Never load-bearing: spawn, key and exit are step-reachable already
			// (the Generate guarantee), so just confirm the lift is not the exit
			// or a lock cell.
			if neck == l.Exit {
				t.Errorf("seed %d: lift placed on the exit", seed)
			}
			if _, locked := l.Locks[neck]; locked {
				t.Errorf("seed %d: lift placed on a locked door", seed)
			}
		}
	}
	if !found {
		t.Skip("no lift generated in scanned seeds")
	}
}
