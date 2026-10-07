package world

import (
	"testing"

	"github.com/danielriddell21/crucible/worldgen"
)

// carve opens the given cells of a fresh solid level as floor.
func carve(t *testing.T, w, h int, cells ...Coord) *Level {
	t.Helper()
	l := newLevel(w, h, 1)
	for _, c := range cells {
		l.Set(c.X, c.Y, TileFloor)
	}
	return l
}

// corridor is a straight run of floor along y = 2 from x = 1 to x = n.
func corridor(n int) []Coord {
	var out []Coord
	for x := 1; x <= n; x++ {
		out = append(out, Coord{X: x, Y: 2})
	}
	return out
}

func TestChebyIsTheLargerAxisDistance(t *testing.T) {
	cases := []struct {
		a, b Coord
		want int
	}{
		{Coord{X: 0, Y: 0}, Coord{X: 0, Y: 0}, 0},
		{Coord{X: 0, Y: 0}, Coord{X: 3, Y: 1}, 3},
		{Coord{X: 0, Y: 0}, Coord{X: 1, Y: 3}, 3},
		{Coord{X: 5, Y: 5}, Coord{X: 2, Y: 4}, 3},
		{Coord{X: 2, Y: 4}, Coord{X: 5, Y: 5}, 3},
		{Coord{X: 0, Y: 5}, Coord{X: 4, Y: 1}, 4},
	}
	for _, c := range cases {
		if got := cheby(c.a, c.b); got != c.want {
			t.Errorf("cheby(%v, %v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestWalkableNeighborsSkipsRockAndEdges(t *testing.T) {
	l := carve(t, 6, 5, Coord{X: 1, Y: 1}, Coord{X: 2, Y: 1}, Coord{X: 1, Y: 2}, Coord{X: 0, Y: 1})
	got := walkableNeighbors(l, Coord{X: 1, Y: 1})
	if len(got) != 3 {
		t.Fatalf("walkableNeighbors = %v, want 3 open sides", got)
	}
	for _, n := range got {
		if !l.At(n.X, n.Y).Walkable() {
			t.Errorf("neighbour %v is not walkable", n)
		}
	}
	if got := walkableNeighbors(l, Coord{X: 0, Y: 1}); len(got) != 1 {
		t.Errorf("a cell on the map edge has neighbours %v, want only the one inside", got)
	}
}

func TestDistanceFieldCountsStepsAlongTheCorridor(t *testing.T) {
	l := carve(t, 8, 5, corridor(5)...)
	d := distanceField(l, Coord{X: 1, Y: 2})
	for x := 1; x <= 5; x++ {
		if got := d[2*l.W+x]; got != x-1 {
			t.Errorf("distance to x=%d is %d, want %d", x, got, x-1)
		}
	}
	if got := d[0]; got != -1 {
		t.Errorf("rock reports distance %d, want -1", got)
	}
}

func TestDistanceFieldOfAnUnwalkableSourceIsAllUnreachable(t *testing.T) {
	l := carve(t, 6, 5, corridor(3)...)
	for _, src := range []Coord{{0, 0}, {-1, 2}, {99, 99}} {
		for i, v := range distanceField(l, src) {
			if v != -1 {
				t.Fatalf("source %v: cell %d has distance %d, want all -1", src, i, v)
			}
		}
	}
}

func TestInOpenBlockNeedsAWholeTwoByTwo(t *testing.T) {
	block := carve(t, 6, 6, Coord{X: 2, Y: 2}, Coord{X: 3, Y: 2}, Coord{X: 2, Y: 3}, Coord{X: 3, Y: 3})
	for _, c := range []Coord{{2, 2}, {3, 2}, {2, 3}, {3, 3}} {
		if !inOpenBlock(block, c) {
			t.Errorf("%v is inside a 2x2 block but inOpenBlock says no", c)
		}
	}
	line := carve(t, 6, 6, corridor(4)...)
	for _, c := range corridor(4) {
		if inOpenBlock(line, c) {
			t.Errorf("%v is in a one-wide corridor but inOpenBlock says yes", c)
		}
	}
}

func TestJunctionMarkersFindATFork(t *testing.T) {
	// A T: a corridor along y=2 with a branch going down at x=3.
	cells := append(corridor(5), Coord{X: 3, Y: 3}, Coord{X: 3, Y: 4})
	l := carve(t, 8, 7, cells...)
	l.Spawn = Coord{X: 1, Y: 2}
	l.Exit = Coord{X: 5, Y: 2}
	d := distanceField(l, l.Exit)

	got := junctionMarkers(l, d)
	if len(got) != 1 {
		t.Fatalf("junctionMarkers = %v, want exactly the fork", got)
	}
	m := got[0]
	if m.Kind != MarkerJunction || m.At != (Coord{X: 3, Y: 2}) {
		t.Errorf("marker = %+v, want a junction at (3,2)", m)
	}
	if len(m.Branches) != 3 {
		t.Errorf("branches = %v, want 3", m.Branches)
	}
	if m.Optimal != (Coord{X: 4, Y: 2}) {
		t.Errorf("optimal branch = %v, want the one toward the exit (4,2)", m.Optimal)
	}
}

func TestJunctionMarkersIgnoreSpawnExitAndOpenRooms(t *testing.T) {
	// A plus whose centre is the spawn point, so it must not be marked.
	plus := []Coord{{3, 3}, {2, 3}, {4, 3}, {3, 2}, {3, 4}}
	l := carve(t, 8, 8, plus...)
	l.Spawn = Coord{X: 3, Y: 3}
	l.Exit = Coord{X: 4, Y: 3}
	if got := junctionMarkers(l, distanceField(l, l.Exit)); len(got) != 0 {
		t.Errorf("markers at the spawn: %v, want none", got)
	}

	room := carve(t, 8, 8, Coord{X: 2, Y: 2}, Coord{X: 3, Y: 2}, Coord{X: 4, Y: 2}, Coord{X: 2, Y: 3}, Coord{X: 3, Y: 3}, Coord{X: 4, Y: 3}, Coord{X: 3, Y: 4})
	room.Spawn = Coord{X: 2, Y: 2}
	room.Exit = Coord{X: 4, Y: 3}
	if got := junctionMarkers(room, distanceField(room, room.Exit)); len(got) != 0 {
		t.Errorf("markers inside an open room: %v, want none", got)
	}
}

func TestDecoyExitMarkerPicksAFloorCellNearTheExit(t *testing.T) {
	l := carve(t, 12, 5, corridor(10)...)
	l.Spawn = Coord{X: 1, Y: 2}
	l.Exit = Coord{X: 9, Y: 2}
	g := worldgen.NewRNG(7)
	m, ok := decoyExitMarker(l, g)
	if !ok {
		t.Fatal("no decoy exit placed")
	}
	if m.Kind != MarkerDecoyExit {
		t.Errorf("kind = %v, want a decoy exit", m.Kind)
	}
	if d := cheby(m.At, l.Exit); d < 2 || d > 5 {
		t.Errorf("decoy at %v is %d from the exit, want 2..5", m.At, d)
	}
	if l.At(m.At.X, m.At.Y) != TileFloor || m.At == l.Spawn {
		t.Errorf("decoy at %v is not a plain floor cell away from the spawn", m.At)
	}

	empty := carve(t, 12, 5, l.Exit)
	empty.Spawn = Coord{X: 0, Y: 0}
	empty.Exit = l.Exit
	if _, ok := decoyExitMarker(empty, g); ok {
		t.Error("placed a decoy with no floor near the exit")
	}
}
