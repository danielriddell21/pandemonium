package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestScentLeadsRoundAWall(t *testing.T) {
	g := openRoom(t, 20, 12)
	for y := 1; y < 9; y++ {
		wall(g, 10, y) // a partition with a gap at the bottom
	}
	g.Player.Pos = Vec2{X: 5.5, Y: 3.5}
	g.trackScent()

	if got := g.scentAt(g.PlayerCell()); got != 0 {
		t.Errorf("scent in the player's cell = %d, want 0", got)
	}
	if got := g.scentAt(world.Coord{X: 10, Y: 3}); got != -1 {
		t.Errorf("scent inside a wall = %d, want -1", got)
	}
	beyond := g.scentAt(world.Coord{X: 12, Y: 3})
	if straight := 12 - 5; beyond <= straight {
		t.Errorf("scent past the partition = %d, want more than the straight line %d", beyond, straight)
	}
	if got := g.scentAt(world.Coord{X: -1, Y: 3}); got != -1 {
		t.Errorf("scent out of bounds = %d, want -1", got)
	}
}

func TestScentRefreshesWhenThePlayerMoves(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Pos = Vec2{X: 5.5, Y: 5.5}
	g.trackScent()
	far := world.Coord{X: 15, Y: 5}
	before := g.scentAt(far)
	g.Player.Pos = Vec2{X: 12.5, Y: 5.5}
	g.trackScent()
	if after := g.scentAt(far); after >= before {
		t.Errorf("scent should shorten as the player nears: %d -> %d", before, after)
	}
}

func TestScentRespectsStepHeight(t *testing.T) {
	g := openRoom(t, 20, 12)
	l := g.World.Level
	// The player stands on a ledge too tall to climb from the floor beside it.
	for y := 1; y < 11; y++ {
		for x := 1; x <= 5; x++ {
			l.SetFloor(x, y, 0.75)
			l.SetCeil(x, y, 1.9)
		}
	}
	g.Player.Pos, g.Player.Z = Vec2{X: 3.5, Y: 5.5}, 0.75
	g.trackScent()
	if got := g.scentAt(world.Coord{X: 8, Y: 5}); got != -1 {
		t.Errorf("a demon cannot climb the ledge, yet the trail reaches it: %d", got)
	}
}

func TestLockedDoorHoldsTheHuntUntilOpened(t *testing.T) {
	g := openRoom(t, 20, 12)
	l := g.World.Level
	for y := 1; y < 11; y++ {
		wall(g, 10, y)
	}
	door := world.Coord{X: 10, Y: 5}
	l.Set(door.X, door.Y, world.TileDoor)
	l.Locks[door] = world.ItemKeyRed
	if g.demonPassable(door) {
		t.Error("a locked, closed door must stop demons")
	}
	g.World.ForceOpenDoor(door.X, door.Y)
	if !g.demonPassable(door) {
		t.Error("an opened keycard door should let demons through")
	}
	if g.demonPassable(world.Coord{X: 10, Y: 3}) || g.demonPassable(world.Coord{X: -4, Y: 0}) {
		t.Error("walls and out-of-bounds cells are never passable")
	}
}

func TestHuntingDemonRoundsACornerAndOpensDoors(t *testing.T) {
	g := openRoom(t, 20, 12)
	l := g.World.Level
	for y := 1; y < 11; y++ {
		wall(g, 10, y)
	}
	door := world.Coord{X: 10, Y: 8}
	l.Set(door.X, door.Y, world.TileDoor)
	g.Player.Pos = Vec2{X: 4.5, Y: 2.5}
	d := demonAt(15.5, 2.5, Pinky)
	d.alert = true
	g.Entities = []Entity{d}

	// Long enough to round the partition, short of reaching the player.
	run(g, Input{}, 60*5)

	if !g.World.Opened(door.X, door.Y) {
		t.Error("a hunting demon should shoulder through an unlocked door")
	}
	if g.Entities[0].Pos.X > 10 {
		t.Errorf("the demon never got past the partition: %+v", g.Entities[0].Pos)
	}
}

func TestIdleDemonDoesNotHunt(t *testing.T) {
	g := openRoom(t, 30, 12)
	g.Player.Pos = Vec2{X: 3.5, Y: 5.5}
	g.Entities = []Entity{demonAt(20.5, 5.5, Melee)}
	start := g.Entities[0].Pos
	run(g, Input{}, 120)
	if g.Entities[0].Pos != start || g.Entities[0].Hunting() {
		t.Error("a demon that has neither seen nor felt the player should stay put")
	}
}

func TestStalkWithoutATrailHoldsStill(t *testing.T) {
	g := openRoom(t, 20, 12)
	d := demonAt(15.5, 5.5, Melee)
	g.Entities = []Entity{d}
	g.stalk(&g.Entities[0], 0, demonSpeed, 1.0/60.0) // no scent laid yet
	if g.Entities[0].Pos != d.Pos {
		t.Error("stalking with no trail should not move the demon")
	}
}

func TestDemonsQueueRatherThanStack(t *testing.T) {
	g := openRoom(t, 30, 5)
	g.Player.Pos = Vec2{X: 3.5, Y: 2.5}
	g.Entities = []Entity{demonAt(8.5, 2.5, Melee), demonAt(9.0, 2.5, Pinky)}
	for i := range g.Entities {
		g.Entities[i].alert = true
	}

	run(g, Input{}, 300)

	if d := dist(g.Entities[0].Pos, g.Entities[1].Pos); d < crowdGap-0.05 {
		t.Errorf("two demons ended %.2f apart, want them queued at least %.2f apart", d, crowdGap)
	}
}

func TestCrowdedOnlyYieldsToDemonsAhead(t *testing.T) {
	g := openRoom(t, 30, 5)
	g.Player.Pos = Vec2{X: 3.5, Y: 2.5}
	g.Entities = []Entity{demonAt(6, 2.5, Melee), demonAt(6.3, 2.5, Melee), newBarrel(Vec2{X: 6.2, Y: 2.5})}
	if !g.crowded(1, Vec2{X: 6.2, Y: 2.5}) {
		t.Error("the demon behind should give way to the one ahead")
	}
	if g.crowded(0, Vec2{X: 5.9, Y: 2.5}) {
		t.Error("the demon in front should not wait for the one behind")
	}
	g.Entities[0].Alive = false
	if g.crowded(1, Vec2{X: 6.2, Y: 2.5}) {
		t.Error("corpses and barrels do not hold up the queue")
	}
}

func TestAdvanceStopsOnTheTarget(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Pos = Vec2{X: 15, Y: 9}
	g.Entities = []Entity{demonAt(5.5, 5.5, Melee)}
	target := Vec2{X: 5.6, Y: 5.5}
	g.advance(&g.Entities[0], 0, target, 10, 1)
	if g.Entities[0].Pos != target {
		t.Errorf("advance overshot its target: %+v", g.Entities[0].Pos)
	}
	g.advance(&g.Entities[0], 0, target, 10, 1)
	if g.Entities[0].Pos != target {
		t.Error("advancing onto the spot already held should be a no-op")
	}
}

func TestCellHelpers(t *testing.T) {
	if got := cellOf(Vec2{X: 3.9, Y: 7.1}); got != (world.Coord{X: 3, Y: 7}) {
		t.Errorf("cellOf = %v", got)
	}
	if got := centreOf(world.Coord{X: 3, Y: 7}); got != (Vec2{X: 3.5, Y: 7.5}) {
		t.Errorf("centreOf = %v", got)
	}
	n := neighbours(world.Coord{X: 2, Y: 2})
	seen := map[world.Coord]bool{}
	for _, c := range n {
		seen[c] = true
		if chebyInt(c, world.Coord{X: 2, Y: 2}) != 1 {
			t.Errorf("neighbour %v is not adjacent", c)
		}
	}
	if len(seen) != 4 {
		t.Errorf("want 4 distinct neighbours, got %d", len(seen))
	}
}
