package world

import "strings"

// Coord is an integer grid coordinate.
type Coord struct {
	X, Y int
}

// Level is a generated map: a row-major grid of tiles plus the points of
// interest needed to play and to reason about it. A Level is produced
// deterministically from Seed (see Generate).
type Level struct {
	Width, Height int
	Tiles         []TileType // row-major: index = y*Width + x, len == Width*Height
	Spawn, Exit   Coord
	Markers       []Marker
	Seed          int64
}

// newLevel allocates a Level of the given size filled entirely with walls.
// Generation then carves floors out of the solid mass.
func newLevel(width, height int, seed int64) *Level {
	tiles := make([]TileType, width*height)
	for i := range tiles {
		tiles[i] = TileWall
	}
	return &Level{
		Width:  width,
		Height: height,
		Tiles:  tiles,
		Seed:   seed,
	}
}

// InBounds reports whether (x, y) lies inside the grid.
func (l *Level) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < l.Width && y < l.Height
}

// At returns the tile at (x, y). Out-of-bounds reads return TileWall so callers
// can treat the world edge as solid without bounds-checking everywhere.
func (l *Level) At(x, y int) TileType {
	if !l.InBounds(x, y) {
		return TileWall
	}
	return l.Tiles[y*l.Width+x]
}

// set writes a tile at (x, y) if in bounds.
func (l *Level) set(x, y int, t TileType) {
	if l.InBounds(x, y) {
		l.Tiles[y*l.Width+x] = t
	}
}

// Solid reports whether (x, y) blocks movement and sight. Walls and the world
// edge are solid; doors are treated as solid here (the simulation decides when a
// specific door has been opened).
func (l *Level) Solid(x, y int) bool {
	switch l.At(x, y) {
	case TileWall, TileDoor:
		return true
	default:
		return false
	}
}

// String renders the grid as ASCII, one row per line. Useful for tests and
// debugging.
func (l *Level) String() string {
	var b strings.Builder
	b.Grow((l.Width + 1) * l.Height)
	for y := range l.Height {
		for x := range l.Width {
			b.WriteRune(l.At(x, y).Rune())
		}
		b.WriteByte('\n')
	}
	return b.String()
}
