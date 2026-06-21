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
	FloorH        []float64  // per-tile floor height in wall units (base 0)
	CeilH         []float64  // per-tile ceiling height in wall units (base 1)
	Spawn, Exit   Coord
	Markers       []Marker
	Items         []Item             // collectibles scattered across the level
	Locks         map[Coord]ItemKind // door cell -> keycard required to open it
	Secrets       []Coord            // cells that count as a hidden find
	Lifts         map[Coord]Lift     // platform tiles that travel between two floors
	Barrels       []Coord            // explosive barrels scattered across the floor
	Hazard        map[Coord]float64  // damaging floor tiles -> health lost per second
	Switches      map[Coord]Switch   // wall switches the player presses with use
	Light         []float64          // per-tile brightness multiplier (1 = full)
	Theme         []uint8            // per-tile wall theme index
	Seed          int64
}

// Lift describes a platform tile that cycles between a low and a high floor,
// serving ledges that plain steps cannot reach. Its tile's static FloorH is Low;
// the simulation animates the platform between the two.
type Lift struct {
	Low, High float64
}

// newLevel allocates a Level of the given size filled entirely with walls, on a
// flat base floor under a flat base ceiling. Generation then carves floors out of
// the solid mass and sculpts the heights.
func newLevel(width, height int, seed int64) *Level {
	tiles := make([]TileType, width*height)
	floors := make([]float64, width*height)
	ceils := make([]float64, width*height)
	light := make([]float64, width*height)
	for i := range tiles {
		tiles[i] = TileWall
		ceils[i] = 1
		light[i] = 1
	}
	return &Level{
		Width:  width,
		Height: height,
		Tiles:  tiles,
		FloorH: floors,
		CeilH:  ceils,
		Light:  light,
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

// Floor returns the floor height at (x, y) in wall units. Out-of-bounds reads
// return 0, matching the solid world edge.
func (l *Level) Floor(x, y int) float64 {
	if !l.InBounds(x, y) {
		return 0
	}
	return l.FloorH[y*l.Width+x]
}

// Ceil returns the ceiling height at (x, y) in wall units. Out-of-bounds reads
// return 0 so the world edge has no gap to slip through.
func (l *Level) Ceil(x, y int) float64 {
	if !l.InBounds(x, y) {
		return 0
	}
	return l.CeilH[y*l.Width+x]
}

// HazardAt returns the health-per-second a tile drains, or 0 if it is safe.
func (l *Level) HazardAt(x, y int) float64 {
	return l.Hazard[Coord{X: x, Y: y}]
}

// LightAt returns the brightness multiplier at (x, y); out of bounds is full.
func (l *Level) LightAt(x, y int) float64 {
	if !l.InBounds(x, y) || len(l.Light) == 0 {
		return 1
	}
	return l.Light[y*l.Width+x]
}

// setFloor / setCeil write heights at (x, y) if in bounds.
func (l *Level) setFloor(x, y int, h float64) {
	if l.InBounds(x, y) {
		l.FloorH[y*l.Width+x] = h
	}
}

func (l *Level) setCeil(x, y int, h float64) {
	if l.InBounds(x, y) {
		l.CeilH[y*l.Width+x] = h
	}
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
	case TileWall, TileDoor, TileSwitch:
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
