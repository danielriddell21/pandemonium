package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const playerRadius = 0.2

const barrelBody = 0.45

func (g *Game) step(pos Vec2, z, dx, dy float64) Vec2 {
	next := pos
	if to := (Vec2{X: pos.X + dx, Y: pos.Y}); !blocked(g.World, to.X, to.Y, z, playerRadius) && !g.barrelBlocks(pos, to) {
		next.X = to.X
	}
	if to := (Vec2{X: next.X, Y: pos.Y + dy}); !blocked(g.World, to.X, to.Y, z, playerRadius) && !g.barrelBlocks(next, to) {
		next.Y = to.Y
	}
	return next
}

func (g *Game) barrelBlocks(from, to Vec2) bool {
	for i := range g.Entities {
		e := &g.Entities[i]
		if e.Kind != Barrel || e.State != Active {
			continue
		}
		// Only closing on a barrel is refused, so a body that somehow overlaps
		// one can always back out.
		if d := dist(to, e.Pos); d < barrelBody && d < dist(from, e.Pos) {
			return true
		}
	}
	return false
}

func blocked(w *World, x, y, z, radius float64) bool {
	minX := int(math.Floor(x - radius))
	maxX := int(math.Floor(x + radius))
	minY := int(math.Floor(y - radius))
	maxY := int(math.Floor(y + radius))
	for ty := minY; ty <= maxY; ty++ {
		for tx := minX; tx <= maxX; tx++ {
			if w.Solid(tx, ty) {
				return true
			}
			f := w.FloorAt(tx, ty)
			if f-z > world.MaxStep {
				return true // step face too tall to climb
			}
			if w.CeilAt(tx, ty)-f < world.MinHeadroom {
				return true // the cell itself is too cramped to stand in
			}
		}
	}
	return false
}
