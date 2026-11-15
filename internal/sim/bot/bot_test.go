package bot

import (
	"math"
	"testing"

	"github.com/danielriddell21/crucible/level"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func room(t *testing.T, w, h int) *sim.Game {
	t.Helper()
	base := level.New(w, h, 1)
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			base.Set(x, y, world.TileFloor)
		}
	}
	base.Spawn = world.Coord{X: 2, Y: h / 2}
	base.Exit = world.Coord{X: w - 3, Y: h / 2}
	g := sim.New(&world.Level{
		Level:    base,
		Locks:    map[world.Coord]world.ItemKind{},
		Hazard:   map[world.Coord]world.HazardCell{},
		Switches: map[world.Coord]world.Switch{},
	})
	g.Entities = nil
	g.Items = nil
	return g
}

func demon(x, y float64, kind sim.EntityKind) sim.Entity {
	return sim.Entity{Pos: sim.Vec2{X: x, Y: y}, Kind: kind, State: sim.Active, Health: 500, Alive: true}
}

func generate(t *testing.T, w, h int, seed int64, arena bool) *world.Level {
	t.Helper()
	l, err := world.Generate(world.Config{Width: w, Height: h, Seed: seed, Arena: arena})
	if err != nil {
		t.Fatalf("seed %d: %v", seed, err)
	}
	return l
}

func play(g *sim.Game, p *Pilot, ticks int) bool {
	for range ticks {
		g.Tick(p.Input(g), 1.0/60.0)
		if g.LevelComplete() {
			return true
		}
	}
	return false
}

func TestRoamerWalksEveryLevelUnopposed(t *testing.T) {
	// With nothing to fight this is a pure test of routing and of the levels:
	// every one must be walkable start to finish, lifts, locks and all.
	for seed := int64(1); seed <= 40; seed++ {
		g := sim.New(generate(t, 32, 24, seed, false))
		g.Entities = nil
		if !play(g, Roamer(false), 60*120) {
			t.Errorf("seed %d: roamer never reached the exit of an empty level (stopped at %v)", seed, g.PlayerCell())
		}
	}
}

func TestRoamerFightsThroughMostLevels(t *testing.T) {
	const seeds = 30
	done := 0
	for seed := int64(1); seed <= seeds; seed++ {
		g := sim.New(generate(t, 32, 24, seed, false))
		if play(g, Roamer(true), 60*180) {
			done++
		}
	}
	if done*10 < seeds*9 {
		t.Errorf("roamer cleared only %d of %d levels", done, seeds)
	}
}

func TestRoamerWaitsForARaisedLift(t *testing.T) {
	// Seed 108 routes over a lift on the way to its keycard: the pilot must
	// wait for the platform rather than give up on the route.
	g := sim.New(generate(t, 48, 32, 108, false))
	g.Entities = nil
	if len(g.World.Level.Lifts) == 0 {
		t.Fatal("seed 108 should carry a lift")
	}
	if !play(g, Roamer(false), 60*90) {
		t.Errorf("roamer did not get past the lift (stopped at %v)", g.PlayerCell())
	}
}

func TestLiftRaised(t *testing.T) {
	g := room(t, 12, 8)
	lift := world.Coord{X: 5, Y: 4}
	g.World.Level.Lifts = map[world.Coord]world.Lift{lift: {Low: 0, High: 0.75}}
	if liftRaised(g, lift) {
		t.Error("a lift at rest on the floor is not raised")
	}
	for range int(60 * 3.6) { // dwell, then all the way up
		g.Tick(sim.Input{}, 1.0/60)
	}
	if !liftRaised(g, lift) {
		t.Error("a lift at the top should read as raised")
	}
	if liftRaised(g, world.Coord{X: 6, Y: 4}) {
		t.Error("an ordinary floor cell is never a raised lift")
	}
}

func TestRoamerFetchesTheKeyBeforeTheExit(t *testing.T) {
	for seed := int64(1); seed <= 20; seed++ {
		l := generate(t, 48, 32, seed, false)
		if len(l.Locks) == 0 {
			continue
		}
		g := sim.New(l)
		key, ok := neededKeyCell(g)
		if !ok {
			t.Fatalf("seed %d: locked level reports no key to fetch", seed)
		}
		route := planRoute(g, false)
		if len(route) == 0 || route[len(route)-1] != key {
			t.Fatalf("seed %d: without the key the route should end at it, ends %v", seed, route[len(route)-1])
		}
		g.Player.Keys = map[world.ItemKind]bool{world.ItemKeyRed: true}
		if _, still := neededKeyCell(g); still {
			t.Errorf("seed %d: a held key should not be fetched again", seed)
		}
		route = planRoute(g, false)
		if route[len(route)-1] != l.Exit {
			t.Errorf("seed %d: with the key the route should end at the exit, ends %v", seed, route[len(route)-1])
		}
		return
	}
	t.Fatal("no locked level in the scanned seeds")
}

func TestRoamerKillsTheBossAndLeaves(t *testing.T) {
	l := generate(t, 48, 32, 5, true)
	g := sim.New(l, sim.WithDepth(world.ArenaEvery-1))
	if c, ok := bossCell(g); !ok || c != l.BossAt {
		t.Fatalf("bossCell = %v, %v; want %v", c, ok, l.BossAt)
	}
	if stop := stops(g, true); stop[len(stop)-1] != l.BossAt {
		t.Errorf("with the exit sealed the pilot should head for the boss, heads for %v", stop)
	}
	if !play(g, Roamer(true), 60*240) {
		t.Fatal("roamer did not finish the boss floor")
	}
	if g.Sealed() {
		t.Error("the level completed with the overlord alive")
	}
}

func TestSteerTowardClosesDistance(t *testing.T) {
	for _, angle := range []float64{0, math.Pi / 3, math.Pi, -math.Pi / 2} {
		g := room(t, 16, 12)
		g.Player.Pos = sim.Vec2{X: 6.5, Y: 6.5}
		g.Player.Angle = angle
		target := sim.Vec2{X: 7.5, Y: 6.5}
		before := dist(g.Player.Pos, target)
		for range 30 {
			g.Tick(steerToward(g, target), 1.0/60.0)
		}
		if after := dist(g.Player.Pos, target); after >= before {
			t.Errorf("angle %.2f: steering widened the gap (%.2f -> %.2f)", angle, before, after)
		}
	}
}

func TestPathIsWalkableStepByStep(t *testing.T) {
	l := generate(t, 32, 24, 7, false)
	g := sim.New(l)
	g.Player.Keys = map[world.ItemKind]bool{world.ItemKeyRed: true}
	route := findPath(g, l.Spawn, l.Exit)
	if len(route) == 0 {
		t.Fatal("no route from spawn to exit")
	}
	if route[0] != l.Spawn || route[len(route)-1] != l.Exit {
		t.Fatalf("route endpoints %v..%v, want %v..%v", route[0], route[len(route)-1], l.Spawn, l.Exit)
	}
	for i := 1; i < len(route); i++ {
		a, b := route[i-1], route[i]
		if abs(a.X-b.X)+abs(a.Y-b.Y) != 1 {
			t.Fatalf("route step %v -> %v is not adjacent", a, b)
		}
		if !l.At(b.X, b.Y).Walkable() {
			t.Fatalf("route crosses unwalkable cell %v", b)
		}
		if l.Floor(b.X, b.Y)-l.Floor(a.X, a.Y) > world.MaxStep+1e-9 {
			t.Fatalf("route climbs too tall a step %v -> %v", a, b)
		}
	}
}

func TestPathRefusesALockedDoorWithoutItsKey(t *testing.T) {
	g := room(t, 16, 8)
	l := g.World.Level
	for y := 1; y < 7; y++ {
		l.Set(8, y, world.TileWall)
	}
	door := world.Coord{X: 8, Y: 4}
	l.Set(door.X, door.Y, world.TileDoor)
	l.Locks[door] = world.ItemKeyRed
	if p := findPath(g, l.Spawn, l.Exit); p != nil {
		t.Errorf("a locked door was routed through without its key: %v", p)
	}
	g.Player.Keys = map[world.ItemKind]bool{world.ItemKeyRed: true}
	if p := findPath(g, l.Spawn, l.Exit); len(p) == 0 {
		t.Error("with the key the door should be routable")
	}
	if p := findPath(g, world.Coord{X: 0, Y: 0}, l.Exit); p != nil {
		t.Error("a path cannot start inside a wall")
	}
}

func TestPathGoesRoundPoolsAndBarrels(t *testing.T) {
	g := room(t, 16, 9)
	l := g.World.Level
	from, to := world.Coord{X: 2, Y: 4}, world.Coord{X: 13, Y: 4}
	l.Hazard[world.Coord{X: 7, Y: 4}] = world.HazardCell{Rate: 8}
	g.Entities = []sim.Entity{{Pos: sim.Vec2{X: 10.5, Y: 4.5}, Kind: sim.Barrel, State: sim.Active, Alive: true}}
	for _, c := range findPath(g, from, to) {
		if l.HazardAt(c.X, c.Y) > 0 {
			t.Errorf("route wades through the pool at %v with dry floor either side", c)
		}
		if c == (world.Coord{X: 10, Y: 4}) {
			t.Error("route walks into the barrel with room to go round")
		}
	}

	// Flood the whole width: now wading is the only way, and it is taken.
	for y := 1; y < 8; y++ {
		l.Hazard[world.Coord{X: 7, Y: y}] = world.HazardCell{Rate: 8}
	}
	if len(findPath(g, from, to)) == 0 {
		t.Error("when a pool is the only way across, the route should wade")
	}
}

func TestForagingDetoursForWhatIsNeeded(t *testing.T) {
	g := room(t, 20, 9)
	g.Items = []sim.ItemState{
		{Kind: world.ItemHealth, Pos: sim.Vec2{X: 4.5, Y: 2.5}},
		{Kind: world.ItemKeyRed, Pos: sim.Vec2{X: 3.5, Y: 4.5}},
	}
	if _, ok := lootNearby(g); ok {
		t.Error("at full health a medkit is not worth a detour, nor is a stray key")
	}
	g.Player.Health = 40
	c, ok := lootNearby(g)
	if !ok || c != (world.Coord{X: 4, Y: 2}) {
		t.Fatalf("hurt, the pilot should go for the medkit: %v %v", c, ok)
	}
	if route := planRoute(g, true); route[0] != g.PlayerCell() || !contains(route, c) {
		t.Errorf("foraging route should pass the medkit: %v", route)
	}
	if contains(planRoute(g, false), c) {
		t.Error("a pilot that is not foraging should walk straight to the exit")
	}
	g.Items[0].Taken = true
	if _, ok := lootNearby(g); ok {
		t.Error("taken items are not loot")
	}
	g.Items = []sim.ItemState{{Kind: world.ItemHealth, Pos: sim.Vec2{X: 17.5, Y: 4.5}}}
	if _, ok := lootNearby(g); ok {
		t.Error("loot beyond reach is not worth leaving the route for")
	}
}

func contains(route []world.Coord, c world.Coord) bool {
	for _, r := range route {
		if r == c {
			return true
		}
	}
	return false
}

func TestWanted(t *testing.T) {
	full := sim.Player{Health: 100, Armor: 100, Bullets: 200, Shells: 50, Rockets: 50}
	spent := sim.Player{Health: 30}
	for _, k := range []world.ItemKind{world.ItemHealth, world.ItemArmor, world.ItemBullets, world.ItemShells, world.ItemRockets} {
		if wanted(full, k) {
			t.Errorf("a full player should not want %v", k)
		}
		if !wanted(spent, k) {
			t.Errorf("a spent player should want %v", k)
		}
	}
	if !wanted(full, world.ItemMega) || !wanted(full, world.ItemBackpack) {
		t.Error("powerups and packs are always wanted")
	}
	if wanted(spent, world.ItemKeyBlue) {
		t.Error("keys are routed for on their own, not foraged")
	}
}

func TestExitSwitchCell(t *testing.T) {
	l := generate(t, 32, 24, 3, false)
	c, ok := exitSwitchCell(l)
	if !ok || l.Switches[c].Action != world.SwitchExit {
		t.Errorf("exitSwitchCell = %v, %v", c, ok)
	}
	if _, ok := exitSwitchCell(room(t, 8, 8).World.Level); ok {
		t.Error("a level without an exit switch should report none")
	}
}

func TestOpenDoorAhead(t *testing.T) {
	g := room(t, 12, 8)
	l := g.World.Level
	l.Set(6, 4, world.TileDoor)
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 4.5}, 0
	var in sim.Input
	openDoorAhead(g, &in)
	if !in.Interact {
		t.Error("facing a closed door the pilot should use it")
	}
	g.World.ForceOpenDoor(6, 4)
	in = sim.Input{}
	openDoorAhead(g, &in)
	if in.Interact {
		t.Error("an open door needs no using")
	}
	sw := world.Coord{X: 5, Y: 0}
	l.Set(sw.X, sw.Y, world.TileSwitch)
	l.Switches[sw] = world.Switch{Action: world.SwitchExit}
	g.Player.Pos, g.Player.Angle = sim.Vec2{X: 5.5, Y: 1.5}, -math.Pi/2
	openDoorAhead(g, &in)
	if !in.Interact {
		t.Error("facing a switch the pilot should throw it")
	}
}

func TestUnstickSidestepsThenReplans(t *testing.T) {
	g := room(t, 12, 8)
	r := &roamer{}
	r.plan(g)
	r.lastPos = g.Player.Pos
	var nudges []float64
	for range replanTicks {
		nudges = append(nudges, r.unstick(g))
	}
	if nudges[0] != 0 || nudges[stuckTicks] == 0 {
		t.Errorf("should wait %d ticks, then sidestep: %v...", stuckTicks, nudges[:stuckTicks+2])
	}
	if nudges[stuckTicks] == nudges[2*stuckTicks] {
		t.Error("the sidestep should alternate direction")
	}
	if r.stuck != 0 {
		t.Errorf("after a full stall the pilot should have replanned and reset: stuck=%d", r.stuck)
	}
	g.Player.Pos.X += 1
	if r.unstick(g) != 0 || r.stuck != 0 {
		t.Error("a pilot that moved is not stuck")
	}
}

func TestRoamerWithNoRouteWalksForward(t *testing.T) {
	g := room(t, 12, 8)
	g.World.Level.Set(g.World.Level.Exit.X, g.World.Level.Exit.Y, world.TileWall)
	if in := Roamer(false).Input(g); in.Forward != 1 {
		t.Errorf("with nowhere to go the roamer should press on, got %+v", in)
	}
}

func TestGeometryHelpers(t *testing.T) {
	if got := cellOf(sim.Vec2{X: 3.9, Y: 7.1}); got != (world.Coord{X: 3, Y: 7}) {
		t.Errorf("cellOf = %v", got)
	}
	if got := centre(world.Coord{X: 3, Y: 7}); got != (sim.Vec2{X: 3.5, Y: 7.5}) {
		t.Errorf("centre = %v", got)
	}
	if d := angleDiff(0.1, 2*math.Pi-0.1); math.Abs(d-0.2) > 1e-9 {
		t.Errorf("angleDiff across the wrap = %v, want 0.2", d)
	}
	if d := angleDiff(-3, 3); math.Abs(d-(2*math.Pi-6)) > 1e-9 {
		t.Errorf("angleDiff the short way round = %v", d)
	}
	if clampF(5, -1, 1) != 1 || clampF(-5, -1, 1) != -1 || clampF(0.3, -1, 1) != 0.3 {
		t.Error("clampF")
	}
	if abs(-4) != 4 || abs(4) != 4 {
		t.Error("abs")
	}
}
