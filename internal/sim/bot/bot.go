// Package bot drives a simulation with simple goal-seeking input: it routes to
// the exit (detouring for keycards), opens doors and switches, and optionally
// fights the demons it meets. The app uses it for the title screen's attract
// loop and the demo capture tool replays it for the documentation clips.
package bot

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// Pilot produces one tick of input at a time for a running game. It keeps just
// enough memory between ticks (the planned route and progress along it) to walk
// a level smoothly; a new level resets it automatically.
type Pilot struct {
	step func(*sim.Game) sim.Input
}

// Input returns the pilot's input for the current tick.
func (p *Pilot) Input(g *sim.Game) sim.Input { return p.step(g) }

// Roamer returns a pilot that walks the shortest route to the exit, detouring to
// grab the keycard that unlocks it, opening doors in its way and throwing the
// exit switch at the end. When fight is set it breaks off to shoot demons within
// reach; otherwise it simply walks the level.
func Roamer(fight bool) *Pilot {
	var (
		level   *world.Level
		route   []world.Coord
		idx     int
		lastPos sim.Vec2
		stuck   int
	)
	plan := func(g *sim.Game) {
		level = g.World.Level
		route = planRoute(g)
		idx = 0
	}
	return &Pilot{step: func(g *sim.Game) sim.Input {
		if g.World.Level != level {
			plan(g)
		}
		if len(route) == 0 {
			return sim.Input{Forward: 1}
		}

		// Break off to fight a demon that has come within reach, so the run
		// actually clears some of the level rather than strolling past them.
		if fight {
			if e, ok := engageDemon(g); ok {
				return engage(g, e)
			}
		}

		// Advance past waypoints we've reached. Use the occupied cell, not exact
		// centring, since a body on a raised step can't always sit dead-centre.
		for idx < len(route)-1 && (g.PlayerCell() == route[idx] || dist(g.Player.Pos, centre(route[idx])) < 0.5) {
			idx++
		}
		target := centre(route[idx])
		if dist(g.Player.Pos, target) > 3 { // teleported (e.g. a death respawn)
			plan(g)
			target = centre(route[idx])
		}

		// If we've barely moved for a while, we're snagged on a corner or step —
		// skip the troublesome waypoint and strafe to slip free.
		var nudge float64
		if dist(g.Player.Pos, lastPos) < 0.02 {
			if stuck++; stuck > 10 {
				if idx < len(route)-1 {
					idx++
					target = centre(route[idx])
				}
				nudge = 1
				stuck = 0
			}
		} else {
			stuck = 0
		}
		lastPos = g.Player.Pos

		// At the end of the route, turn to the exit switch and throw it.
		if idx == len(route)-1 {
			if sw, ok := exitSwitchCell(g.World.Level); ok {
				target = centre(sw)
			}
		}

		// Steer toward the waypoint by sliding on both axes, not just walking the
		// way we face — this rounds corners and stepped corridors without snagging.
		in := steerToward(g, target)
		in.Strafe += nudge
		openDoorAhead(g, &in)
		if fight {
			attackIfThreatened(g, &in)
		}
		return in
	}}
}

// Hunter returns a pilot that works through the level's demons one after
// another: it shoots whatever it can see, and navigates toward the nearest
// survivor when none is in sight, so a clip or attract run is a sustained fight
// rather than a single kill.
func Hunter() *Pilot {
	var (
		lvl    *world.Level
		path   []world.Coord
		idx    int
		toCell world.Coord
	)
	return &Pilot{step: func(g *sim.Game) sim.Input {
		e, ok := nearestDemon(g)
		if !ok {
			return sim.Input{} // all clear — hold while the last kill settles
		}
		// Fight anything in sight; otherwise close the gap toward it.
		if losClear(g.World, g.Player.Pos, e.Pos) && dist(e.Pos, g.Player.Pos) <= 6 {
			return engage(g, e)
		}

		ec := world.Coord{X: int(e.Pos.X), Y: int(e.Pos.Y)}
		if g.World.Level != lvl || ec != toCell || idx >= len(path) {
			lvl, toCell = g.World.Level, ec
			path = bfsOpen(lvl, g.PlayerCell(), ec)
			idx = 0
		}
		if len(path) == 0 {
			return sim.Input{Forward: 1}
		}
		target := centre(path[idx])
		for dist(g.Player.Pos, target) < 0.35 && idx < len(path)-1 {
			idx++
			target = centre(path[idx])
		}
		diff := angleDiff(math.Atan2(target.Y-g.Player.Pos.Y, target.X-g.Player.Pos.X), g.Player.Angle)
		var in sim.Input
		if math.Abs(diff) > 0.05 {
			in.TurnDelta = clampF(diff, -0.18, 0.18)
		}
		if math.Abs(diff) < 0.7 {
			in.Forward = 1
		}
		openDoorAhead(g, &in)
		return in
	}}
}

// engageDemon returns the nearest living demon within a short fighting range and
// clear line of sight, so the roamer can deal with it before moving on.
func engageDemon(g *sim.Game) (sim.Entity, bool) {
	const engageRange = 5.0
	best := math.Inf(1)
	var found sim.Entity
	ok := false
	for _, e := range g.Entities {
		if !e.Alive {
			continue
		}
		d := dist(e.Pos, g.Player.Pos)
		if d > engageRange || !losClear(g.World, g.Player.Pos, e.Pos) {
			continue
		}
		if d < best {
			best, found, ok = d, e, true
		}
	}
	return found, ok
}

// engage turns to face a demon, closes to firing distance and shoots it.
func engage(g *sim.Game, e sim.Entity) sim.Input {
	diff := angleDiff(math.Atan2(e.Pos.Y-g.Player.Pos.Y, e.Pos.X-g.Player.Pos.X), g.Player.Angle)
	var in sim.Input
	if math.Abs(diff) > 0.04 {
		in.TurnDelta = clampF(diff, -0.16, 0.16)
	}
	if math.Abs(diff) < 0.4 && dist(e.Pos, g.Player.Pos) > 2.2 {
		in.Forward = 1
	}
	attackIfThreatened(g, &in)
	return in
}

// attackIfThreatened sets Attack when a living demon is just ahead in range.
func attackIfThreatened(g *sim.Game, in *sim.Input) {
	dir := g.Player.Dir()
	for _, e := range g.Entities {
		if !e.Alive {
			continue
		}
		dx, dy := e.Pos.X-g.Player.Pos.X, e.Pos.Y-g.Player.Pos.Y
		d := math.Hypot(dx, dy)
		if d == 0 || d > 2.9 {
			continue
		}
		if (dx/d)*dir.X+(dy/d)*dir.Y < 0.9 {
			continue
		}
		if losClear(g.World, g.Player.Pos, e.Pos) {
			in.Attack = true
			return
		}
	}
}

// nearestDemon returns the closest living demon to the player.
func nearestDemon(g *sim.Game) (sim.Entity, bool) {
	best := math.Inf(1)
	var found sim.Entity
	ok := false
	for _, e := range g.Entities {
		if !e.Alive {
			continue
		}
		if d := dist(e.Pos, g.Player.Pos); d < best {
			best, found, ok = d, e, true
		}
	}
	return found, ok
}

// losClear samples the segment between two points and reports whether it crosses
// any solid tile. It is a coarse "can I shoot it" answer for steering, not the
// simulation's exact sight rule.
func losClear(w *sim.World, a, b sim.Vec2) bool {
	steps := int(dist(a, b)/0.1) + 1
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := a.X + (b.X-a.X)*t
		y := a.Y + (b.Y-a.Y)*t
		if w.Solid(int(math.Floor(x)), int(math.Floor(y))) {
			return false
		}
	}
	return true
}

func centre(c world.Coord) sim.Vec2 { return sim.Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5} }
func dist(a, b sim.Vec2) float64    { return math.Hypot(a.X-b.X, a.Y-b.Y) }

func angleDiff(target, current float64) float64 {
	d := math.Mod(target-current+math.Pi, 2*math.Pi)
	if d < 0 {
		d += 2 * math.Pi
	}
	return d - math.Pi
}

func clampF(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }
