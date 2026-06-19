package world

import "testing"

func TestExitSwitchPlacedBesideExit(t *testing.T) {
	withSwitch := 0
	for seed := int64(0); seed < 60; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if !l.HasExitSwitch() {
			continue
		}
		withSwitch++
		for c, s := range l.Switches {
			if s.Action != SwitchExit {
				continue
			}
			if l.At(c.X, c.Y) != TileSwitch {
				t.Errorf("seed %d: exit switch cell %v is not a TileSwitch", seed, c)
			}
			if !l.Solid(c.X, c.Y) {
				t.Errorf("seed %d: switch %v should be solid", seed, c)
			}
			if cheby(c, l.Exit) != 1 {
				t.Errorf("seed %d: exit switch %v not adjacent to exit %v", seed, c, l.Exit)
			}
		}
	}
	if withSwitch < 50 {
		t.Errorf("only %d/60 levels got an exit switch; expected nearly all", withSwitch)
	}
}

func TestExitStaysReachable(t *testing.T) {
	// The exit floor tile must remain reachable with doors open even though the
	// adjacent switch wall is solid.
	for seed := int64(0); seed < 100; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if !reachable(l, l.Spawn, l.Exit, blocksWalls(l)) {
			t.Errorf("seed %d: exit unreachable", seed)
		}
	}
}
