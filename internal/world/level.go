package world

import "strings"

type Coord struct {
	X, Y int
}

type Level struct {
	Width, Height int
	Tiles         []TileType
	FloorH        []float64
	CeilH         []float64
	Spawn, Exit   Coord
	Markers       []Marker
	Items         []Item
	Locks         map[Coord]ItemKind
	Secrets       []Coord
	Lifts         map[Coord]Lift
	Barrels       []Coord
	Hazard        map[Coord]HazardCell
	Switches      map[Coord]Switch
	Light         []float64
	Theme         []uint8
	WallTop       []float64
	Sky           []bool
	Seed          int64
}

type Lift struct {
	Low, High float64
}

func newLevel(width, height int, seed int64) *Level {
	tiles := make([]TileType, width*height)
	floors := make([]float64, width*height)
	ceils := make([]float64, width*height)
	light := make([]float64, width*height)
	wallTop := make([]float64, width*height) // 0 everywhere: all walls full-height
	sky := make([]bool, width*height)
	for i := range tiles {
		tiles[i] = TileWall
		ceils[i] = 1
		light[i] = 1
	}
	return &Level{
		Width:   width,
		Height:  height,
		Tiles:   tiles,
		FloorH:  floors,
		CeilH:   ceils,
		Light:   light,
		WallTop: wallTop,
		Sky:     sky,
		Seed:    seed,
	}
}

func (l *Level) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < l.Width && y < l.Height
}

func (l *Level) At(x, y int) TileType {
	if !l.InBounds(x, y) {
		return TileWall
	}
	return l.Tiles[y*l.Width+x]
}

func (l *Level) Floor(x, y int) float64 {
	if !l.InBounds(x, y) {
		return 0
	}
	return l.FloorH[y*l.Width+x]
}

func (l *Level) Ceil(x, y int) float64 {
	if !l.InBounds(x, y) {
		return 0
	}
	return l.CeilH[y*l.Width+x]
}

func (l *Level) HazardAt(x, y int) float64 {
	return l.Hazard[Coord{X: x, Y: y}].Rate
}

func (l *Level) HazardKindAt(x, y int) HazardKind {
	return l.Hazard[Coord{X: x, Y: y}].Kind
}

func (l *Level) WallTopAt(x, y int) float64 {
	if !l.InBounds(x, y) || len(l.WallTop) == 0 {
		return 0
	}
	return l.WallTop[y*l.Width+x]
}

func (l *Level) LightAt(x, y int) float64 {
	if !l.InBounds(x, y) || len(l.Light) == 0 {
		return 1
	}
	return l.Light[y*l.Width+x]
}

func (l *Level) SkyAt(x, y int) bool {
	if !l.InBounds(x, y) || len(l.Sky) == 0 {
		return false
	}
	return l.Sky[y*l.Width+x]
}

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

func (l *Level) set(x, y int, t TileType) {
	if l.InBounds(x, y) {
		l.Tiles[y*l.Width+x] = t
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
