package world

import (
	"testing"

	"github.com/danielriddell21/crucible/worldgen"
)

func dryField(l *Level, lock *Coord) worldgen.Field {
	solid := func(c Coord) bool {
		if _, pool := l.Hazard[c]; pool || !l.At(c.X, c.Y).Walkable() {
			return true
		}
		return lock != nil && c == *lock
	}
	return worldgen.FloodDist(l.W, l.H, l.Spawn, solid, climbable(l))
}

func TestPoolsNeverCutOffTheExitOrAKey(t *testing.T) {
	pools := 0
	for _, dims := range [][2]int{{48, 32}, {32, 24}} {
		for seed := int64(1); seed <= 150; seed++ {
			l, err := Generate(Config{Width: dims[0], Height: dims[1], Seed: seed})
			if err != nil {
				t.Fatal(err)
			}
			if len(l.Hazard) == 0 {
				continue
			}
			pools++
			if dryField(l, nil).At(l.Exit) < 0 {
				t.Errorf("%v seed %d: the only way to the exit wades through a pool", dims, seed)
			}
			for lock, key := range l.Locks {
				at, ok := keyCell(l, key)
				if !ok {
					t.Fatalf("%v seed %d: lock without a key", dims, seed)
				}
				if dryField(l, &lock).At(at) < 0 {
					t.Errorf("%v seed %d: the only way to the %v wades through a pool", dims, seed, key)
				}
				if l.HazardAt(at.X, at.Y) > 0 {
					t.Errorf("%v seed %d: the %v lies in a pool", dims, seed, key)
				}
			}
		}
	}
	if pools < 100 {
		t.Errorf("only %d of 300 levels carry a pool; the dry-route rule should not starve them", pools)
	}
}

func TestWetRefusesACellThatSeversTheLevel(t *testing.T) {
	l := newLevel(16, 16, 1)
	for x := 1; x < 15; x++ {
		l.Set(x, 8, TileFloor) // one corridor: every cell of it is the only way
	}
	l.Spawn, l.Exit = Coord{X: 1, Y: 8}, Coord{X: 14, Y: 8}
	cell := HazardCell{Rate: nukageRate}
	if wet(l, Coord{X: 7, Y: 8}, cell) {
		t.Error("a pool on the only corridor must be refused")
	}
	if len(l.Hazard) != 0 {
		t.Error("a refused cell must not stay wet")
	}
	if floodHazard(l, Coord{X: 7, Y: 8}, HazardNukage, nukageRate) {
		t.Error("a flood that cannot start should report failure")
	}

	// Widen the corridor: now there is a dry way round.
	for x := 5; x < 10; x++ {
		l.Set(x, 9, TileFloor)
	}
	if !wet(l, Coord{X: 7, Y: 8}, cell) {
		t.Error("with a way round, the cell may flood")
	}
	if !dryRoutes(l) {
		t.Error("the level must still have a dry route")
	}
}

func TestDryRoutesGuardsTheKey(t *testing.T) {
	l := newLevel(16, 16, 1)
	for x := 1; x < 15; x++ {
		l.Set(x, 8, TileFloor)
	}
	for y := 3; y < 8; y++ {
		l.Set(4, y, TileFloor) // a side passage up to the key
	}
	l.Spawn, l.Exit = Coord{X: 1, Y: 8}, Coord{X: 14, Y: 8}
	lock := Coord{X: 10, Y: 8}
	l.Set(lock.X, lock.Y, TileDoor)
	l.Locks[lock] = ItemKeyRed
	l.Items = []Item{{Kind: ItemKeyRed, At: Coord{X: 4, Y: 3}}}
	if !dryRoutes(l) {
		t.Fatal("a dry level should pass")
	}
	l.Hazard[Coord{X: 4, Y: 5}] = HazardCell{Rate: lavaRate, Kind: HazardLava}
	if dryRoutes(l) {
		t.Error("a pool across the only passage to the key must fail the check")
	}
	l.Hazard = map[Coord]HazardCell{}
	l.Items = nil
	if dryRoutes(l) {
		t.Error("a lock with no key cannot be satisfied")
	}
}

func TestBarrelsOnlyStandWhereTheyCanBeWalkedAround(t *testing.T) {
	total := 0
	for seed := int64(1); seed <= 80; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range l.Barrels {
			total++
			if !ringOpen(l, b) {
				t.Errorf("seed %d: barrel at %v is hemmed in and would block the way", seed, b)
			}
		}
	}
	if total < 80 {
		t.Errorf("only %d barrels across 80 levels; placement has become too strict", total)
	}
}

func TestRingOpen(t *testing.T) {
	l := newLevel(8, 8, 1)
	for y := 2; y <= 4; y++ {
		for x := 2; x <= 4; x++ {
			l.Set(x, y, TileFloor)
		}
	}
	if !ringOpen(l, Coord{X: 3, Y: 3}) {
		t.Error("the middle of a 3x3 room is open all round")
	}
	if ringOpen(l, Coord{X: 2, Y: 3}) {
		t.Error("a cell against a wall is not open all round")
	}
}

func TestAlarmPanelIsMountedOnAReachableWall(t *testing.T) {
	mounted := 0
	for seed := int64(1); seed <= 80; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		c, ok := l.AlarmPanel()
		if !ok {
			continue
		}
		mounted++
		if l.At(c.X, c.Y) != TileSwitch || !l.Solid(c.X, c.Y) {
			t.Errorf("seed %d: panel %v is not a solid switch tile", seed, c)
		}
		if cheby(c, l.Spawn) < 4 || cheby(c, l.Exit) < 3 {
			t.Errorf("seed %d: panel %v crowds the spawn or the exit", seed, c)
		}
		faces := 0
		for _, n := range neighbors4(c) {
			if l.InBounds(n.X, n.Y) && l.At(n.X, n.Y).Walkable() {
				faces++
				if !reachable(l, l.Spawn, n, blocksWalls(l)) {
					t.Errorf("seed %d: nobody can stand in front of panel %v", seed, c)
				}
			}
		}
		if faces != 1 {
			t.Errorf("seed %d: panel %v has %d open faces, want 1", seed, c, faces)
		}
		if !l.HasExitSwitch() {
			t.Errorf("seed %d: the alarm panel displaced the exit switch", seed)
		}
	}
	if mounted < 76 {
		t.Errorf("only %d/80 levels got an alarm panel", mounted)
	}
}

func TestAlarmPanelDoesNotDisturbTheRestOfTheLevel(t *testing.T) {
	a, _ := Generate(Config{Width: 40, Height: 28, Seed: 11})
	b, _ := Generate(Config{Width: 40, Height: 28, Seed: 11})
	pa, _ := a.AlarmPanel()
	pb, _ := b.AlarmPanel()
	if pa != pb || a.String() != b.String() {
		t.Error("panel placement must be deterministic")
	}
}

func TestPanelWallRules(t *testing.T) {
	l := newLevel(16, 16, 1)
	for x := 1; x < 15; x++ {
		l.Set(x, 8, TileFloor)
	}
	l.Spawn, l.Exit = Coord{X: 1, Y: 8}, Coord{X: 14, Y: 8}
	if !panelWall(l, Coord{X: 7, Y: 7}) {
		t.Error("a wall with one floor face, clear of spawn and exit, should take a panel")
	}
	if panelWall(l, Coord{X: 2, Y: 7}) {
		t.Error("a wall beside the spawn should not")
	}
	if panelWall(l, Coord{X: 13, Y: 7}) {
		t.Error("a wall beside the exit should not")
	}
	if panelWall(l, Coord{X: 7, Y: 8}) {
		t.Error("a floor cell is not a wall")
	}
	l.Set(7, 8, TileDoor)
	if panelWall(l, Coord{X: 7, Y: 7}) {
		t.Error("a panel must not face a doorway")
	}
	empty := newLevel(16, 16, 1)
	placeAlarmPanel(empty)
	if _, ok := empty.AlarmPanel(); ok {
		t.Error("solid rock has nowhere to mount a panel")
	}
}

func TestArenaIsABossFloorWithCover(t *testing.T) {
	l, err := Generate(Config{Width: 48, Height: 32, Seed: 5, Arena: true})
	if err != nil {
		t.Fatal(err)
	}
	if !l.Boss || l.At(l.BossAt.X, l.BossAt.Y) != TileFloor {
		t.Fatalf("arena should name a floor cell for its boss: boss=%v at %v", l.Boss, l.BossAt)
	}
	if cheby(l.BossAt, l.Spawn) < 10 {
		t.Errorf("boss at %v starts on top of the spawn %v", l.BossAt, l.Spawn)
	}
	pillars := 0
	for y := 3; y < l.H-3; y++ {
		for x := 3; x < l.W-3; x++ {
			if l.At(x, y) == TileWall {
				pillars++
				if y == l.Spawn.Y {
					t.Errorf("pillar at %d,%d blocks the centre line", x, y)
				}
			}
		}
	}
	if pillars != 16 {
		t.Errorf("want four 2x2 pillars (16 wall cells), got %d", pillars)
	}
	if _, ok := l.AlarmPanel(); ok {
		t.Error("the boss floor has no alarm to silence")
	}
	for _, b := range l.Barrels {
		if !ringOpen(l, b) {
			t.Errorf("arena barrel %v is not walkable around", b)
		}
	}
	maze, _ := Generate(Config{Width: 48, Height: 32, Seed: 5})
	if maze.Boss {
		t.Error("an ordinary level is not a boss floor")
	}
}
