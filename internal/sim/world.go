package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// World wraps an immutable generated level with the mutable runtime state the
// simulation needs on top of it — which doors have been opened and where each
// lift platform currently sits. The underlying level is never modified.
type World struct {
	Level  *world.Level
	opened map[world.Coord]bool
	clock  float64 // drives lift platforms; advanced by Tick
}

// NewWorld wraps a generated level for simulation.
func NewWorld(l *world.Level) *World {
	return &World{Level: l, opened: make(map[world.Coord]bool)}
}

// Tick advances the world's moving parts (lift platforms) by dt seconds.
func (w *World) Tick(dt float64) {
	w.clock += dt
}

// Lift platform timing: rest at each end, then travel between floors.
const (
	liftDwell  = 2.0 // seconds parked at the bottom or top
	liftTravel = 1.5 // seconds spent moving between floors
)

// FloorAt returns the effective floor height at (x, y): a lift's current
// platform height, or the level's sculpted floor everywhere else.
func (w *World) FloorAt(x, y int) float64 {
	if lf, ok := w.Level.Lifts[world.Coord{X: x, Y: y}]; ok {
		return liftHeight(lf, w.clock)
	}
	return w.Level.Floor(x, y)
}

// CeilAt returns the ceiling height at (x, y).
func (w *World) CeilAt(x, y int) float64 {
	return w.Level.Ceil(x, y)
}

// HazardAt returns the health-per-second the floor at (x, y) drains, or 0.
func (w *World) HazardAt(x, y int) float64 {
	return w.Level.HazardAt(x, y)
}

// HazardKindAt returns the kind of hazard on the floor at (x, y).
func (w *World) HazardKindAt(x, y int) world.HazardKind {
	return w.Level.HazardKindAt(x, y)
}

// liftHeight is a lift platform's height at time t: dwell low, rise, dwell high,
// sink, repeating. Pure in t, so the cycle is deterministic from the tick count.
func liftHeight(lf world.Lift, t float64) float64 {
	period := 2 * (liftDwell + liftTravel)
	p := math.Mod(t, period)
	switch {
	case p < liftDwell:
		return lf.Low
	case p < liftDwell+liftTravel:
		return lf.Low + (lf.High-lf.Low)*(p-liftDwell)/liftTravel
	case p < 2*liftDwell+liftTravel:
		return lf.High
	default:
		return lf.High - (lf.High-lf.Low)*(p-2*liftDwell-liftTravel)/liftTravel
	}
}

// Solid reports whether the cell at integer (x, y) blocks movement. Walls and
// the world edge are always solid; a door is solid until it has been opened.
func (w *World) Solid(x, y int) bool {
	switch w.Level.At(x, y) {
	case world.TileWall, world.TileSwitch:
		return true
	case world.TileDoor:
		return !w.opened[world.Coord{X: x, Y: y}]
	default:
		return false
	}
}

// IsDoor reports whether (x, y) is a door tile.
func (w *World) IsDoor(x, y int) bool {
	return w.Level.At(x, y) == world.TileDoor
}

// Opened reports whether the door at (x, y) has been opened.
func (w *World) Opened(x, y int) bool {
	return w.opened[world.Coord{X: x, Y: y}]
}

// Lock returns the keycard required to open the door at (x, y), and whether the
// door is locked at all.
func (w *World) Lock(x, y int) (world.ItemKind, bool) {
	k, ok := w.Level.Locks[world.Coord{X: x, Y: y}]
	return k, ok
}

// OpenDoor marks the door at (x, y) open, provided any lock on it is satisfied by
// hasKey. It reports whether this call changed the state (i.e. the cell was a
// still-closed, unlocked-or-keyed door).
func (w *World) OpenDoor(x, y int, hasKey func(world.ItemKind) bool) bool {
	if w.Level.At(x, y) != world.TileDoor {
		return false
	}
	c := world.Coord{X: x, Y: y}
	if w.opened[c] {
		return false
	}
	if key, locked := w.Level.Locks[c]; locked && !hasKey(key) {
		return false
	}
	w.opened[c] = true
	return true
}

// ForceOpenDoor opens the door at (x, y) regardless of any lock — used by remote
// switches. It reports whether the state changed.
func (w *World) ForceOpenDoor(x, y int) bool {
	if w.Level.At(x, y) != world.TileDoor {
		return false
	}
	c := world.Coord{X: x, Y: y}
	if w.opened[c] {
		return false
	}
	w.opened[c] = true
	return true
}
