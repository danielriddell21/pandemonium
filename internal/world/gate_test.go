package world

import "testing"

// firstGatedLevel returns a generated level that has at least one locked door,
// scanning seeds until it finds one.
func firstGatedLevel(t *testing.T) *Level {
	t.Helper()
	for seed := int64(0); seed < 200; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if len(l.Locks) > 0 {
			return l
		}
	}
	t.Skip("no gated level found in scanned seeds")
	return nil
}

func TestLockedDoorGatesTheExit(t *testing.T) {
	l := firstGatedLevel(t)
	for lock := range l.Locks {
		// The door must sit on a door tile.
		if l.At(lock.X, lock.Y) != TileDoor {
			t.Errorf("lock %v is not on a door tile", lock)
		}
		// Sealing the locked cell must cut the exit off entirely: it is a true
		// chokepoint, not a door with a way around it.
		if reachable(l, l.Spawn, l.Exit, lockedSolid(l, lock)) {
			t.Errorf("exit still reachable with lock %v sealed; door does not gate", lock)
		}
	}
}

func TestKeyReachableWithoutItsDoor(t *testing.T) {
	l := firstGatedLevel(t)
	for lock, key := range l.Locks {
		cell, ok := keyCell(l, key)
		if !ok {
			t.Fatalf("no %v item placed for lock %v", key, lock)
		}
		if !reachable(l, l.Spawn, cell, lockedSolid(l, lock)) {
			t.Errorf("%v at %v is unreachable without crossing its own door %v", key, cell, lock)
		}
	}
}

func TestGateIsDeterministic(t *testing.T) {
	cfg := Config{Width: 48, Height: 32, Seed: 13}
	a, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Locks) != len(b.Locks) {
		t.Fatalf("lock count differs: %d vs %d", len(a.Locks), len(b.Locks))
	}
	for c, k := range a.Locks {
		if b.Locks[c] != k {
			t.Errorf("lock at %v differs: %v vs %v", c, k, b.Locks[c])
		}
	}
}
