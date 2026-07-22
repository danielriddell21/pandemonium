package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const playerRadius = 0.2

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
