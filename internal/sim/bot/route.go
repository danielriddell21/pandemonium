package bot

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// planRoute builds the cells to walk: detour through the exit's keycard first (so
// the locked door will open when reached), then on to the exit. Doors are treated
// as passable for routing; the pilot opens them as it arrives.
func planRoute(g *sim.Game) []world.Coord {
	l := g.World.Level
	start := g.PlayerCell()
	var route []world.Coord
	if key, ok := firstKeyCell(l); ok {
		if seg := bfsOpen(l, start, key); len(seg) > 0 {
			route = seg
			start = key
		}
	}
	if seg := bfsOpen(l, start, l.Exit); len(seg) > 0 {
		if len(route) > 0 {
			seg = seg[1:] // drop the shared cell where the segments meet
		}
		route = append(route, seg...)
	}
	return route
}

// firstKeyCell returns the location of the first keycard item on the level.
func firstKeyCell(l *world.Level) (world.Coord, bool) {
	for _, it := range l.Items {
		if it.Kind.IsKey() {
			return it.At, true
		}
	}
	return world.Coord{}, false
}

// exitSwitchCell returns the cell of the level's exit switch, if it has one.
func exitSwitchCell(l *world.Level) (world.Coord, bool) {
	for c, s := range l.Switches {
		if s.Action == world.SwitchExit {
			return c, true
		}
	}
	return world.Coord{}, false
}

// openDoorAhead issues Interact when an unopened door or a switch sits directly
// in front of the player, mirroring how the simulation decides what the player
// can act on.
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

// steerToward turns to face a target and moves toward it on both the forward and
// strafe axes, so the pilot slides around corners and up stepped corridors
// instead of stalling when it can't walk in a dead-straight line.
func steerToward(g *sim.Game, target sim.Vec2) sim.Input {
	toX, toY := target.X-g.Player.Pos.X, target.Y-g.Player.Pos.Y
	desired := math.Atan2(toY, toX)
	var in sim.Input
	in.TurnDelta = clampF(angleDiff(desired, g.Player.Angle), -0.2, 0.2)
	dir := g.Player.Dir()
	fwd := toX*dir.X + toY*dir.Y    // component along facing
	str := toX*(-dir.Y) + toY*dir.X // component along the strafe axis
	in.Forward = clampF(fwd, -1, 1)
	in.Strafe = clampF(str, -1, 1)
	return in
}

// bfsOpen finds the shortest cell path from src to dst over a level, treating
// every door as passable (walls and the world edge are the only obstacles), so
// the pilot routes through locked doors it intends to open.
func bfsOpen(l *world.Level, src, dst world.Coord) []world.Coord {
	passable := func(c world.Coord) bool {
		return l.InBounds(c.X, c.Y) && l.At(c.X, c.Y).Walkable()
	}
	// climbable mirrors the simulation's step rule: a body can step up at most
	// world.MaxStep and drop any distance, so the pilot only routes where it can
	// actually walk.
	climbable := func(from, to world.Coord) bool {
		return l.Floor(to.X, to.Y)-l.Floor(from.X, from.Y) <= world.MaxStep+1e-9
	}
	if !passable(src) || !passable(dst) {
		return nil
	}
	prev := map[world.Coord]world.Coord{src: src}
	queue := []world.Coord{src}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == dst {
			break
		}
		for _, n := range neighbours(c) {
			if !passable(n) || !climbable(c, n) {
				continue
			}
			if _, seen := prev[n]; seen {
				continue
			}
			prev[n] = c
			queue = append(queue, n)
		}
	}
	if _, ok := prev[dst]; !ok {
		return nil
	}
	var rev []world.Coord
	for c := dst; c != src; c = prev[c] {
		rev = append(rev, c)
	}
	rev = append(rev, src)
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

func neighbours(c world.Coord) [4]world.Coord {
	return [4]world.Coord{
		{X: c.X + 1, Y: c.Y}, {X: c.X - 1, Y: c.Y},
		{X: c.X, Y: c.Y + 1}, {X: c.X, Y: c.Y - 1},
	}
}
