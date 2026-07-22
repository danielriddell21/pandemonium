package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

type World struct {
	Level  *world.Level
	opened map[world.Coord]bool
	clock  float64
}

func NewWorld(l *world.Level) *World {
	return &World{Level: l, opened: make(map[world.Coord]bool)}
}

func (w *World) Tick(dt float64) {
	w.clock += dt
}

const (
	liftDwell  = 2.0
	liftTravel = 1.5
)

func (w *World) FloorAt(x, y int) float64 {
	if lf, ok := w.Level.Lifts[world.Coord{X: x, Y: y}]; ok {
		return liftHeight(lf, w.clock)
	}
	return w.Level.Floor(x, y)
}

func (w *World) CeilAt(x, y int) float64 {
	return w.Level.Ceil(x, y)
}

func (w *World) HazardAt(x, y int) float64 {
	return w.Level.HazardAt(x, y)
}

func (w *World) HazardKindAt(x, y int) world.HazardKind {
	return w.Level.HazardKindAt(x, y)
}

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

func (w *World) IsDoor(x, y int) bool {
	return w.Level.At(x, y) == world.TileDoor
}

func (w *World) Opened(x, y int) bool {
	return w.opened[world.Coord{X: x, Y: y}]
}

func (w *World) Lock(x, y int) (world.ItemKind, bool) {
	k, ok := w.Level.Locks[world.Coord{X: x, Y: y}]
	return k, ok
}

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
