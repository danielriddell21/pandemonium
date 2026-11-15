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

const (
	stuckTicks   = 20
	replanTicks  = 120
	refreshTicks = 90
	stuckEps     = 0.01
)

type roamer struct {
	fight   bool
	level   *world.Level
	route   []world.Coord
	idx     int
	age     int
	lastPos sim.Vec2
	stuck   int
	gunner  gunner
}

func (r *roamer) plan(g *sim.Game) {
	r.level = g.World.Level
	r.route = planRoute(g, r.fight)
	r.idx = 0
	r.age = 0
}

func (r *roamer) tick(g *sim.Game) sim.Input {
	// The world moves under the plan: keys and loot get taken, a quarry
	// wanders, health drops. Re-reading it on a short timer covers them all.
	if r.age++; g.World.Level != r.level || r.age >= refreshTicks {
		r.plan(g)
	}
	if len(r.route) == 0 {
		return sim.Input{Forward: 1}
	}
	if r.fight {
		if in, ok := r.gunner.fight(g); ok {
			return in
		}
	}

	target := r.nextTarget(g)
	if liftRaised(g, r.route[r.idx]) {
		r.stuck = 0
		return sim.Input{} // the platform comes back down; skipping ahead would loop
	}
	nudge := r.unstick(g)

	// At the end of the route, turn to the exit switch and throw it.
	if r.idx == len(r.route)-1 && r.route[r.idx] == g.World.Level.Exit {
		if sw, ok := exitSwitchCell(g.World.Level); ok {
			target = centre(sw)
		}
	}

	// Steer toward the waypoint by sliding on both axes, not just walking the
	// way we face — this rounds corners and stepped corridors without snagging.
	in := steerToward(g, target)
	in.Strafe += nudge
	openDoorAhead(g, &in)
	return in
}

func (r *roamer) nextTarget(g *sim.Game) sim.Vec2 {
	for r.idx < len(r.route)-1 && (g.PlayerCell() == r.route[r.idx] || dist(g.Player.Pos, centre(r.route[r.idx])) < 0.5) {
		r.idx++
	}
	target := centre(r.route[r.idx])
	if dist(g.Player.Pos, target) > 3 { // off the route: a death respawn, or a fight that wandered
		r.plan(g)
		target = centre(r.route[r.idx])
	}
	return target
}

func (r *roamer) unstick(g *sim.Game) float64 {
	moved := dist(g.Player.Pos, r.lastPos) >= stuckEps
	r.lastPos = g.Player.Pos
	if moved {
		r.stuck = 0
		return 0
	}
	r.stuck++
	if r.stuck >= replanTicks {
		r.plan(g)
		r.stuck = 0
		return 0
	}
	if r.stuck < stuckTicks {
		return 0
	}
	// Sidestep one way, then the other, until something gives.
	if (r.stuck/stuckTicks)%2 == 1 {
		return 1
	}
	return -1
}

func liftRaised(g *sim.Game, next world.Coord) bool {
	if _, lift := g.World.Level.Lifts[next]; !lift {
		return false
	}
	return g.World.FloorAt(next.X, next.Y)-g.Player.Z > world.MaxStep
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

func cellOf(p sim.Vec2) world.Coord {
	return world.Coord{X: int(math.Floor(p.X)), Y: int(math.Floor(p.Y))}
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
