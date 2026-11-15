package world

import (
	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/level"
)

type (
	Coord = geom.Coord
	rect  = geom.Rect
)

type Lift = level.Lift

type Level struct {
	*level.Level
	Markers  []Marker
	Items    []Item
	Locks    map[Coord]ItemKind
	Secrets  []Coord
	Barrels  []Coord
	Hazard   map[Coord]HazardCell
	Switches map[Coord]Switch
	Boss     bool
	BossAt   Coord
}

func newLevel(width, height int, seed int64) *Level {
	return &Level{
		Level:    level.New(width, height, seed),
		Locks:    map[Coord]ItemKind{},
		Hazard:   map[Coord]HazardCell{},
		Switches: map[Coord]Switch{},
	}
}

func (l *Level) Solid(x, y int) bool {
	switch l.At(x, y) {
	case TileWall, TileDoor, TileSwitch:
		return true
	default:
		return false
	}
}

func (l *Level) HazardAt(x, y int) float64 {
	return l.Hazard[Coord{X: x, Y: y}].Rate
}

func (l *Level) HazardKindAt(x, y int) HazardKind {
	return l.Hazard[Coord{X: x, Y: y}].Kind
}

func (l *Level) WallTopAt(x, y int) float64 {
	return l.WallTop(x, y)
}
