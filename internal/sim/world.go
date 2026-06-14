package sim

import "github.com/danielriddell21/pandemonium/internal/world"

// World wraps an immutable generated level with the mutable runtime state the
// simulation needs on top of it — currently which doors have been opened. The
// underlying level is never modified.
type World struct {
	Level  *world.Level
	opened map[world.Coord]bool
}

// NewWorld wraps a generated level for simulation.
func NewWorld(l *world.Level) *World {
	return &World{Level: l, opened: make(map[world.Coord]bool)}
}

// Solid reports whether the cell at integer (x, y) blocks movement. Walls and
// the world edge are always solid; a door is solid until it has been opened.
func (w *World) Solid(x, y int) bool {
	switch w.Level.At(x, y) {
	case world.TileWall:
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
