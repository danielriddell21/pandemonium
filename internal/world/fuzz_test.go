package world

import (
	"errors"
	"testing"
)

func FuzzGenerate(f *testing.F) {
	f.Add(48, 32, int64(1), false)
	f.Add(16, 16, int64(2), false)
	f.Add(64, 40, int64(3), false)
	f.Add(40, 28, int64(5), true)
	f.Add(31, 19, int64(-7), false)

	f.Fuzz(func(t *testing.T, width, height int, seed int64, arena bool) {
		// Keep dimensions in a sane, bounded range: generation clamps up to its
		// minimum anyway, and an upper bound keeps each iteration quick.
		width = clampInt(width, minDimension, 80)
		height = clampInt(height, minDimension, 80)

		l, err := Generate(Config{Width: width, Height: height, Seed: seed, Arena: arena})
		if err != nil {
			if errors.Is(err, ErrUnreachable) {
				return // a documented, non-panic outcome
			}
			t.Fatalf("unexpected error: %v", err)
		}

		if !l.InBounds(l.Spawn.X, l.Spawn.Y) || !l.InBounds(l.Exit.X, l.Exit.Y) {
			t.Fatalf("spawn %v or exit %v out of bounds (%dx%d)", l.Spawn, l.Exit, l.Width, l.Height)
		}
		if !reachable(l, l.Spawn, l.Exit, blocksWalls(l)) {
			t.Fatal("exit not reachable from spawn once doors are open")
		}
		if !keysReachable(l) {
			t.Fatal("a keycard is locked behind the door it opens")
		}
		const eps = 1e-9
		for y := range l.Height {
			for x := range l.Width {
				if l.At(x, y).Walkable() && l.Ceil(x, y)-l.Floor(x, y) < MinHeadroom-eps {
					t.Fatalf("cell %d,%d lacks headroom: floor %.2f ceil %.2f", x, y, l.Floor(x, y), l.Ceil(x, y))
				}
			}
		}
		for c := range l.Hazard {
			if !l.At(c.X, c.Y).Walkable() {
				t.Fatalf("hazard at %v is not walkable", c)
			}
		}

		// Determinism: the same config must reproduce an identical level.
		l2, err := Generate(Config{Width: width, Height: height, Seed: seed, Arena: arena})
		if err != nil {
			t.Fatalf("second generation errored: %v", err)
		}
		if l.String() != l2.String() {
			t.Fatal("generation is not deterministic for a fixed config")
		}
	})
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
