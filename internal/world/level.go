package world

import (
	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/level"
)

// Coord is a tile-grid cell and rect an axis-aligned span of tiles, shared
// with the family through crucible/geom.
type (
	Coord = geom.Coord
	rect  = geom.Rect
)

// Lift is a moving platform's low and high rest heights, from crucible/level.
type Lift = level.Lift

// Level is a crucible/level spatial world plus pandemonium's gameplay layer:
// the markers that annotate junctions and decoys, the items and the locks
// that gate them, the secrets, the barrels, the hazard field, and the
// switches.
type Level struct {
	*level.Level
	Markers  []Marker
	Items    []Item
	Locks    map[Coord]ItemKind
	Secrets  []Coord
	Barrels  []Coord
	Hazard   map[Coord]HazardCell
	Switches map[Coord]Switch
}

// newLevel returns an aggregate wrapping a fresh solid-rock crucible/level,
// with the gameplay maps initialised.
func newLevel(width, height int, seed int64) *Level {
	return &Level{
		Level:    level.New(width, height, seed),
		Locks:    map[Coord]ItemKind{},
		Hazard:   map[Coord]HazardCell{},
		Switches: map[Coord]Switch{},
	}
}

// Solid reports whether the cell blocks movement and sight. Unlike the
// engine's walkability rule, a closed door and a wall-mounted switch both
// block.
func (l *Level) Solid(x, y int) bool {
	switch l.At(x, y) {
	case TileWall, TileDoor, TileSwitch:
		return true
	default:
		return false
	}
}

// HazardAt returns the damage rate of the hazard on the cell, or 0.
func (l *Level) HazardAt(x, y int) float64 {
	return l.Hazard[Coord{X: x, Y: y}].Rate
}

// HazardKindAt returns the kind of hazard on the cell, or the zero kind.
func (l *Level) HazardKindAt(x, y int) HazardKind {
	return l.Hazard[Coord{X: x, Y: y}].Kind
}

// WallTopAt returns the half-wall height at the cell, or 0 for a full wall.
func (l *Level) WallTopAt(x, y int) float64 {
	return l.WallTop(x, y)
}
