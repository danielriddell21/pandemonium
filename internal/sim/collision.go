package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// playerRadius keeps the player from clipping into walls; collision treats the
// player as a small axis-aligned box of this half-extent.
const playerRadius = 0.2

// resolveMove slides a body at pos (standing at height z) by (dx, dy) against
// solid tiles and unclimbable height changes, axis by axis so the player grazes
// walls smoothly instead of sticking. It returns the new position.
func resolveMove(w *World, pos Vec2, z, dx, dy float64) Vec2 {
	next := pos
	if !blocked(w, pos.X+dx, pos.Y, z, playerRadius) {
		next.X = pos.X + dx
	}
	if !blocked(w, next.X, pos.Y+dy, z, playerRadius) {
		next.Y = pos.Y + dy
	}
	return next
}

// blocked reports whether a box of the given radius centred at world-space
// (x, y), standing at height z, overlaps any tile it cannot occupy: a solid
// tile, a floor rising more than a step above its feet, or a ceiling too low to
// stand under.
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
			if w.CeilAt(tx, ty)-math.Max(f, z) < world.MinHeadroom {
				return true // not enough room to stand
			}
		}
	}
	return false
}
