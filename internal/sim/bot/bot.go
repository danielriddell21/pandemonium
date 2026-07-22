package bot

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

type Pilot struct {
	step func(*sim.Game) sim.Input
}

func (p *Pilot) Input(g *sim.Game) sim.Input { return p.step(g) }

func Roamer(fight bool) *Pilot {
	r := &roamer{fight: fight}
	return &Pilot{step: r.tick}
}

type roamer struct {
	fight   bool
	level   *world.Level
	route   []world.Coord
	idx     int
	lastPos sim.Vec2
	stuck   int
}

func (r *roamer) plan(g *sim.Game) {
	r.level = g.World.Level
	r.route = planRoute(g)
	r.idx = 0
}

func (r *roamer) tick(g *sim.Game) sim.Input {
	if g.World.Level != r.level {
		r.plan(g)
	}
	if len(r.route) == 0 {
		return sim.Input{Forward: 1}
	}

	// Break off to fight a demon that has come within reach, so the run
	// actually clears some of the level rather than strolling past them.
	if r.fight {
		if e, ok := engageDemon(g); ok {
			return engage(g, e)
		}
	}

	target := r.nextTarget(g)
	nudge := r.unstick(g, &target)

	// At the end of the route, turn to the exit switch and throw it.
	if r.idx == len(r.route)-1 {
		if sw, ok := exitSwitchCell(g.World.Level); ok {
			target = centre(sw)
		}
	}

	// Steer toward the waypoint by sliding on both axes, not just walking the
	// way we face — this rounds corners and stepped corridors without snagging.
	in := steerToward(g, target)
	in.Strafe += nudge
	openDoorAhead(g, &in)
	if r.fight {
		attackIfThreatened(g, &in)
	}
	return in
}

func (r *roamer) nextTarget(g *sim.Game) sim.Vec2 {
	for r.idx < len(r.route)-1 && (g.PlayerCell() == r.route[r.idx] || dist(g.Player.Pos, centre(r.route[r.idx])) < 0.5) {
		r.idx++
	}
	target := centre(r.route[r.idx])
	if dist(g.Player.Pos, target) > 3 { // teleported (e.g. a death respawn)
		r.plan(g)
		target = centre(r.route[r.idx])
	}
	return target
}

func (r *roamer) unstick(g *sim.Game, target *sim.Vec2) float64 {
	var nudge float64
	if dist(g.Player.Pos, r.lastPos) < 0.02 {
		if r.stuck++; r.stuck > 10 {
			if r.idx < len(r.route)-1 {
				r.idx++
				*target = centre(r.route[r.idx])
			}
			nudge = 1
			r.stuck = 0
		}
	} else {
		r.stuck = 0
	}
	r.lastPos = g.Player.Pos
	return nudge
}

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

func clampF(v, lo, hi float64) float64 { return max(lo, min(hi, v)) }
