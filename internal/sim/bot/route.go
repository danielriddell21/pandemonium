package bot

import (
	"container/heap"
	"math"
	"slices"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	hazardCost = 25
	barrelCost = 40
	lootReach  = 10
)

func planRoute(g *sim.Game, forage bool) []world.Coord {
	start := g.PlayerCell()
	var route []world.Coord
	for _, stop := range stops(g, forage) {
		seg := findPath(g, start, stop)
		if len(seg) == 0 {
			continue
		}
		if len(route) > 0 {
			seg = seg[1:] // drop the shared cell where the segments meet
		}
		route = append(route, seg...)
		start = stop
	}
	return route
}

func stops(g *sim.Game, forage bool) []world.Coord {
	var out []world.Coord
	if forage {
		if c, ok := lootNearby(g); ok {
			out = append(out, c)
		}
		// A sealed exit is no destination: go and find what seals it.
		if c, ok := bossCell(g); ok {
			return append(out, c)
		}
	}
	if key, ok := neededKeyCell(g); ok {
		out = append(out, key)
	}
	return append(out, g.World.Level.Exit)
}

func bossCell(g *sim.Game) (world.Coord, bool) {
	for _, e := range g.Entities {
		if e.Kind == sim.Boss && e.Alive {
			return cellOf(e.Pos), true
		}
	}
	return world.Coord{}, false
}

func lootNearby(g *sim.Game) (world.Coord, bool) {
	best, bestLen := world.Coord{}, lootReach+1
	here := g.PlayerCell()
	for _, it := range g.Items {
		c := cellOf(it.Pos)
		if it.Taken || !wanted(g.Player, it.Kind) || abs(c.X-here.X)+abs(c.Y-here.Y) >= bestLen {
			continue
		}
		if n := len(findPath(g, here, c)); n > 1 && n <= bestLen {
			best, bestLen = c, n-1
		}
	}
	return best, bestLen <= lootReach
}

func wanted(p sim.Player, k world.ItemKind) bool {
	switch k {
	case world.ItemHealth:
		return p.Health < 0.7*sim.MaxHealth
	case world.ItemArmor:
		return p.Armor < 0.5*sim.MaxArmor
	case world.ItemBullets:
		return p.Bullets < p.MaxBullets()/2
	case world.ItemShells:
		return p.Shells < p.MaxShells()/2
	case world.ItemRockets:
		return p.Rockets < p.MaxRockets()/2
	default:
		return !k.IsKey() // powerups and packs are always worth the steps
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func neededKeyCell(g *sim.Game) (world.Coord, bool) {
	l := g.World.Level
	for _, key := range l.Locks {
		if g.Player.HasKey(key) {
			continue
		}
		for _, it := range l.Items {
			if it.Kind == key {
				return it.At, true
			}
		}
	}
	return world.Coord{}, false
}

func exitSwitchCell(l *world.Level) (world.Coord, bool) {
	for c, s := range l.Switches {
		if s.Action == world.SwitchExit {
			return c, true
		}
	}
	return world.Coord{}, false
}

func openDoorAhead(g *sim.Game, in *sim.Input) {
	dir := g.Player.Dir()
	tx := int(math.Floor(g.Player.Pos.X + dir.X*0.9))
	ty := int(math.Floor(g.Player.Pos.Y + dir.Y*0.9))
	if _, ok := g.World.Level.SwitchAt(tx, ty); ok {
		in.Interact = true
	}
	if g.World.IsDoor(tx, ty) && !g.World.Opened(tx, ty) {
		in.Interact = true
	}
}

func steerToward(g *sim.Game, target sim.Vec2) sim.Input {
	toX, toY := target.X-g.Player.Pos.X, target.Y-g.Player.Pos.Y
	desired := math.Atan2(toY, toX)
	var in sim.Input
	in.TurnDelta = clampF(angleDiff(desired, g.Player.Angle), -maxTurn, maxTurn)
	dir := g.Player.Dir()
	fwd := toX*dir.X + toY*dir.Y    // component along facing
	str := toX*(-dir.Y) + toY*dir.X // component along the strafe axis
	in.Forward = clampF(fwd, -1, 1)
	in.Strafe = clampF(str, -1, 1)
	return in
}

type node struct {
	at   world.Coord
	cost int
}

type frontier []node

func (f frontier) Len() int           { return len(f) }
func (f frontier) Less(i, j int) bool { return f[i].cost < f[j].cost }
func (f frontier) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f *frontier) Push(x any)        { *f = append(*f, x.(node)) }
func (f *frontier) Pop() any {
	old := *f
	n := old[len(old)-1]
	*f = old[:len(old)-1]
	return n
}

type terrain struct {
	g       *sim.Game
	barrels map[world.Coord]bool
}

func survey(g *sim.Game) terrain {
	t := terrain{g: g, barrels: make(map[world.Coord]bool)}
	for _, e := range g.Entities {
		if e.Kind == sim.Barrel && e.Alive {
			t.barrels[cellOf(e.Pos)] = true // solid: cheaper to walk round than to stop and burst
		}
	}
	return t
}

func (t terrain) passable(c world.Coord) bool {
	l := t.g.World.Level
	if !l.InBounds(c.X, c.Y) || !l.At(c.X, c.Y).Walkable() {
		return false
	}
	if l.At(c.X, c.Y) != world.TileDoor {
		return true
	}
	key, locked := l.Locks[c]
	return !locked || t.g.Player.HasKey(key)
}

func (t terrain) step(from, to world.Coord) (int, bool) {
	l := t.g.World.Level
	// This mirrors the simulation's step rule: a body can step up at most
	// world.MaxStep and drop any distance, so the pilot only routes where it
	// can actually walk.
	if !t.passable(to) || l.Floor(to.X, to.Y)-l.Floor(from.X, from.Y) > world.MaxStep+1e-9 {
		return 0, false
	}
	cost := 1
	if len(l.Hazard) > 0 && l.HazardAt(to.X, to.Y) > 0 {
		cost += hazardCost
	}
	if len(t.barrels) > 0 && t.barrels[to] {
		cost += barrelCost
	}
	return cost, true
}

func findPath(g *sim.Game, src, dst world.Coord) []world.Coord {
	t := survey(g)
	if !t.passable(src) || !t.passable(dst) {
		return nil
	}
	// Dijkstra rather than a plain flood: a pool costs far more than a floor
	// tile, so the route wades through one only when nothing else leads there.
	l := g.World.Level
	at := func(c world.Coord) int { return c.Y*l.W + c.X }
	cost := make([]int, l.W*l.H)
	prev := make([]world.Coord, l.W*l.H)
	seen := make([]bool, l.W*l.H)
	seen[at(src)] = true
	open := &frontier{{at: src}}
	for open.Len() > 0 {
		cur := heap.Pop(open).(node)
		if cur.at == dst {
			break
		}
		if cur.cost > cost[at(cur.at)] {
			continue
		}
		for _, n := range neighbours(cur.at) {
			step, ok := t.step(cur.at, n)
			if !ok {
				continue
			}
			c := cur.cost + step
			if i := at(n); !seen[i] || c < cost[i] {
				seen[i], cost[i], prev[i] = true, c, cur.at
				heap.Push(open, node{at: n, cost: c})
			}
		}
	}
	if !seen[at(dst)] {
		return nil
	}
	var rev []world.Coord
	for c := dst; c != src; c = prev[at(c)] {
		rev = append(rev, c)
	}
	rev = append(rev, src)
	slices.Reverse(rev)
	return rev
}

func neighbours(c world.Coord) [4]world.Coord {
	return [4]world.Coord{
		{X: c.X + 1, Y: c.Y},
		{X: c.X - 1, Y: c.Y},
		{X: c.X, Y: c.Y + 1},
		{X: c.X, Y: c.Y - 1},
	}
}
