package world

import (
	"testing"
)

func TestGenerateConnectivity(t *testing.T) {
	sizes := []struct {
		w, h int
	}{
		{16, 16},
		{32, 24},
		{48, 30},
		{64, 48},
		{80, 60},
	}
	// For each size, many seeds must all yield an exit reachable from spawn.
	for _, s := range sizes {
		for seed := int64(0); seed < 200; seed++ {
			l, err := Generate(Config{Width: s.w, Height: s.h, Seed: seed})
			if err != nil {
				t.Fatalf("%dx%d seed=%d: %v", s.w, s.h, seed, err)
			}
			if l.At(l.Spawn.X, l.Spawn.Y) != TileSpawn {
				t.Errorf("%dx%d seed=%d: spawn tile not marked", s.w, s.h, seed)
			}
			if l.At(l.Exit.X, l.Exit.Y) != TileExit {
				t.Errorf("%dx%d seed=%d: exit tile not marked", s.w, s.h, seed)
			}
			// The exit must be reachable once doors are open, and every keycard
			// must be obtainable without first crossing the door it unlocks.
			if !reachable(l, l.Spawn, l.Exit, blocksWalls(l)) {
				t.Errorf("%dx%d seed=%d: exit unreachable from spawn\n%s", s.w, s.h, seed, l)
			}
			if !keysReachable(l) {
				t.Errorf("%dx%d seed=%d: a keycard is unreachable without its own door\n%s", s.w, s.h, seed, l)
			}
		}
	}
}

func TestGenerateDeterminism(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{"small", Config{Width: 24, Height: 24, Seed: 1}},
		{"wide", Config{Width: 64, Height: 32, Seed: 7}},
		{"large", Config{Width: 80, Height: 60, Seed: 123456}},
		{"zero seed", Config{Width: 40, Height: 30, Seed: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := Generate(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			b, err := Generate(tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			if a.String() != b.String() {
				t.Errorf("same seed produced different grids:\n--- a ---\n%s--- b ---\n%s", a, b)
			}
			if a.Spawn != b.Spawn || a.Exit != b.Exit {
				t.Errorf("spawn/exit differ: a=(%v,%v) b=(%v,%v)", a.Spawn, a.Exit, b.Spawn, b.Exit)
			}
			if !markersEqual(a.Markers, b.Markers) {
				t.Errorf("markers differ between identical seeds: %d vs %d", len(a.Markers), len(b.Markers))
			}
		})
	}
}

func TestGenerateDifferentSeedsDiffer(t *testing.T) {
	a, err := Generate(Config{Width: 48, Height: 32, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(Config{Width: 48, Height: 32, Seed: 2})
	if err != nil {
		t.Fatal(err)
	}
	if a.String() == b.String() {
		t.Error("different seeds produced identical grids")
	}
}

func TestGenerateNormalizesTinyDimensions(t *testing.T) {
	l, err := Generate(Config{Width: 4, Height: 4, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if l.Width < minDimension || l.Height < minDimension {
		t.Errorf("tiny dimensions not normalized: got %dx%d", l.Width, l.Height)
	}
}

func markersEqual(a, b []Marker) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Kind != b[i].Kind || a[i].At != b[i].At || a[i].Optimal != b[i].Optimal {
			return false
		}
		if len(a[i].Branches) != len(b[i].Branches) {
			return false
		}
		for j := range a[i].Branches {
			if a[i].Branches[j] != b[i].Branches[j] {
				return false
			}
		}
	}
	return true
}
