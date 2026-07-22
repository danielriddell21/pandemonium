package world

import "testing"

func TestIsArenaLevel(t *testing.T) {
	for _, n := range []int{5, 10, 15} {
		if !IsArenaLevel(n) {
			t.Errorf("level %d should be an arena", n)
		}
	}
	for _, n := range []int{0, 1, 4, 6, 9} {
		if IsArenaLevel(n) {
			t.Errorf("level %d should not be an arena", n)
		}
	}
}

func TestArenaIsOneOpenReachableRoom(t *testing.T) {
	l, err := Generate(Config{Width: 48, Height: 32, Seed: 5, Arena: true})
	if err != nil {
		t.Fatalf("arena generation failed: %v", err)
	}
	if !reachable(l, l.Spawn, l.Exit, blocksWalls(l)) {
		t.Fatal("arena exit not reachable from spawn")
	}
	if !l.HasExitSwitch() {
		t.Error("arena should end on an exit switch")
	}

	// A single open room: the spawn's whole row between spawn and exit is floor.
	for x := l.Spawn.X; x <= l.Exit.X; x++ {
		if !l.At(x, l.Spawn.Y).Walkable() {
			t.Fatalf("expected open floor across the arena at %d,%d", x, l.Spawn.Y)
		}
	}

	// The cache is present and the room is open to the sky.
	hasMega := false
	for _, it := range l.Items {
		if it.Kind == ItemMega {
			hasMega = true
		}
	}
	if !hasMega {
		t.Error("arena should hold a megasphere cache")
	}
	if !l.SkyAt(l.W/2, l.H/2) {
		t.Error("arena should be open to the sky")
	}
}

func TestArenaDeterministic(t *testing.T) {
	cfg := Config{Width: 48, Height: 32, Seed: 5, Arena: true}
	a, _ := Generate(cfg)
	b, _ := Generate(cfg)
	if a.String() != b.String() {
		t.Error("arena layout differs between identical configs")
	}
	if len(a.Items) != len(b.Items) {
		t.Errorf("arena item count differs: %d vs %d", len(a.Items), len(b.Items))
	}
}

func TestArenaFlagOffIsOrdinaryMaze(t *testing.T) {
	// Without the flag the same seed must produce the usual multi-room maze, not
	// the single open room.
	maze, _ := Generate(Config{Width: 48, Height: 32, Seed: 5})
	arena, _ := Generate(Config{Width: 48, Height: 32, Seed: 5, Arena: true})
	if maze.String() == arena.String() {
		t.Error("arena flag should change the layout")
	}
}
