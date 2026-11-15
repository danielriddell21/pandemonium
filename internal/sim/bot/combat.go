package bot

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

const (
	sightRange   = 7.0
	bossRange    = 12.0
	retreatRange = 3.0
	rocketSafe   = 4.0
	laneWidth    = 0.6
	laneStep     = 0.05
	barrelDanger = 2.6
	strafeTicks  = 45
	aimSlack     = 0.25
	maxTurn      = 0.2
	sideProbe    = 0.7
	ignoreTicks  = 90
)

type gunner struct {
	side    float64
	swap    int
	ignore  int
	lastPos sim.Vec2
}

func (c *gunner) fight(g *sim.Game) (sim.Input, bool) {
	if c.ignore > 0 {
		c.ignore--
		return sim.Input{}, false
	}
	e, ok := threat(g)
	if !ok {
		return sim.Input{}, false
	}
	to := sim.Vec2{X: e.Pos.X - g.Player.Pos.X, Y: e.Pos.Y - g.Player.Pos.Y}
	d := math.Hypot(to.X, to.Y)
	diff := angleDiff(math.Atan2(to.Y, to.X), g.Player.Angle)

	// A rocket that would burst on a step or a lintel in our own face is no
	// option; fall back to something that shoots straight.
	weapon := pickWeapon(g.Player, e, d, rocketLane(g, to, d))
	in := sim.Input{TurnDelta: clampF(diff, -maxTurn, maxTurn)}
	if weapon != g.Player.Weapon {
		in.SelectWeapon = int(weapon) + 1
	}

	rng, arcCos := weapon.Reach()
	// A barrel in the line of fire would take the shot and burst in our face:
	// step back out of its blast first, or leave the fight if there is no room.
	barrel := barrelInLine(g, rng, arcCos)
	if barrel && !c.clear(g, -1, 0) {
		c.ignore = ignoreTicks // walk on rather than dither between the fight and the route
		return sim.Input{}, false
	}
	switch {
	case barrel, d < retreatRange && weapon != sim.Fists && c.clear(g, -1, 0):
		in.Forward = -1 // nothing out-walks a retreating player but a pinky
	case d > rng:
		in.Forward = 1
	}
	in.Strafe = c.sidestep(g, e.Pos)

	aimed := math.Cos(diff) >= arcCos+(1-arcCos)*aimSlack
	if weapon == g.Player.Weapon && d <= rng && aimed && !barrel {
		in.Attack = true
	}
	return in, true
}

func (c *gunner) sidestep(g *sim.Game, target sim.Vec2) float64 {
	if c.side == 0 {
		c.side = 1
	}
	// Keep moving across the line of fire so aimed fireballs miss; turn about
	// on a timer, or at once when a wall or a pool stops the slide.
	stalled := dist(g.Player.Pos, c.lastPos) < stuckEps
	c.lastPos = g.Player.Pos
	if c.swap--; c.swap <= 0 || stalled || !c.open(g, c.side, target) {
		c.side = -c.side
		c.swap = strafeTicks
	}
	if !c.open(g, c.side, target) {
		return 0
	}
	return c.side
}

func (c *gunner) open(g *sim.Game, side float64, target sim.Vec2) bool {
	// Sliding behind a corner would drop the target and hand control back to
	// the route, which walks straight out again.
	return c.clear(g, 0, side) && losClear(g.World, probe(g, 0, side), target)
}

func (c *gunner) clear(g *sim.Game, forward, strafe float64) bool {
	p := probe(g, forward, strafe)
	cx, cy := int(math.Floor(p.X)), int(math.Floor(p.Y))
	return !g.World.Solid(cx, cy) && g.World.HazardAt(cx, cy) == 0
}

func probe(g *sim.Game, forward, strafe float64) sim.Vec2 {
	dir := g.Player.Dir()
	return sim.Vec2{
		X: g.Player.Pos.X + (dir.X*forward-dir.Y*strafe)*sideProbe,
		Y: g.Player.Pos.Y + (dir.Y*forward+dir.X*strafe)*sideProbe,
	}
}

func rocketLane(g *sim.Game, to sim.Vec2, d float64) bool {
	z := g.EyeZ()
	for s := laneStep; s < min(d, rocketSafe); s += laneStep {
		x, y := int(math.Floor(g.Player.Pos.X+to.X/d*s)), int(math.Floor(g.Player.Pos.Y+to.Y/d*s))
		if g.World.Solid(x, y) || z <= g.World.FloorAt(x, y) || z >= g.World.CeilAt(x, y) {
			return false
		}
	}
	// Anything standing beside the lane sets the rocket off as it passes.
	for _, e := range g.Entities {
		if !e.Alive {
			continue
		}
		ex, ey := e.Pos.X-g.Player.Pos.X, e.Pos.Y-g.Player.Pos.Y
		along := (ex*to.X + ey*to.Y) / d
		if across := math.Abs(ex*to.Y-ey*to.X) / d; along > 0 && along < rocketSafe && across < laneWidth {
			return false
		}
	}
	return true
}

func threat(g *sim.Game) (sim.Entity, bool) {
	best := math.Inf(1)
	var found sim.Entity
	ok := false
	for _, e := range g.Entities {
		if !e.Alive || e.Kind == sim.Barrel {
			continue
		}
		d := dist(e.Pos, g.Player.Pos)
		reach := sightRange
		if e.Kind == sim.Boss {
			reach = bossRange // it opens fire from across the room, so answer from there
		}
		if d == 0 || d > reach || d >= best || !losClear(g.World, g.Player.Pos, e.Pos) {
			continue
		}
		best, found, ok = d, e, true
	}
	return found, ok
}

func pickWeapon(p sim.Player, e sim.Entity, d float64, lane bool) sim.WeaponKind {
	shotgunRange, _ := sim.Shotgun.Reach()
	switch {
	case p.Berserk && d < 1.5:
		return sim.Fists
	case p.Rockets > 0 && d > rocketSafe && lane:
		return sim.RocketLauncher
	case p.Shells > 0 && d <= shotgunRange:
		return sim.Shotgun
	case p.Bullets > 0 && (e.Kind == sim.Baron || d < retreatRange):
		return sim.Chaingun
	case p.Bullets > 0:
		return sim.Pistol
	case p.Shells > 0:
		return sim.Shotgun
	default:
		return sim.Fists
	}
}

func barrelInLine(g *sim.Game, rng, arcCos float64) bool {
	dir := g.Player.Dir()
	for _, e := range g.Entities {
		if !e.Alive || e.Kind != sim.Barrel {
			continue
		}
		dx, dy := e.Pos.X-g.Player.Pos.X, e.Pos.Y-g.Player.Pos.Y
		d := math.Hypot(dx, dy)
		if d == 0 || d > min(rng, barrelDanger) {
			continue
		}
		if (dx/d)*dir.X+(dy/d)*dir.Y >= arcCos && losClear(g.World, g.Player.Pos, e.Pos) {
			return true
		}
	}
	return false
}
