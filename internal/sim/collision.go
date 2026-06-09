package sim

import "math"

// playerRadius keeps the player from clipping into walls; collision treats the
// player as a small axis-aligned box of this half-extent.
const playerRadius = 0.2

// resolveMove slides a body at pos by (dx, dy) against solid tiles, axis by axis
// so the player grazes walls smoothly instead of sticking. It returns the new
// position.
func resolveMove(w *World, pos Vec2, dx, dy float64) Vec2 {
	next := pos
	if !blocked(w, pos.X+dx, pos.Y, playerRadius) {
		next.X = pos.X + dx
	}
	if !blocked(w, next.X, pos.Y+dy, playerRadius) {
		next.Y = pos.Y + dy
	}
	return next
}

// blocked reports whether a box of the given radius centred at world-space
// (x, y) overlaps any solid tile.
func blocked(w *World, x, y, radius float64) bool {
	minX := int(math.Floor(x - radius))
	maxX := int(math.Floor(x + radius))
	minY := int(math.Floor(y - radius))
	maxY := int(math.Floor(y + radius))
	for ty := minY; ty <= maxY; ty++ {
		for tx := minX; tx <= maxX; tx++ {
			if w.Solid(tx, ty) {
				return true
			}
		}
	}
	return false
}
