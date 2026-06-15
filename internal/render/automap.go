package render

import (
	"image/color"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// automap layout and colours.
const (
	automapMaxW  = 200
	automapMaxH  = 140
	automapInset = 8
)

var (
	mapPanel  = color.RGBA{R: 14, G: 12, B: 16, A: 255}
	mapWall   = color.RGBA{R: 96, G: 90, B: 78, A: 255}
	mapFloor  = color.RGBA{R: 46, G: 42, B: 36, A: 255}
	mapDoor   = color.RGBA{R: 150, G: 110, B: 70, A: 255}
	mapExit   = color.RGBA{R: 60, G: 200, B: 90, A: 255}
	mapPlayer = color.RGBA{R: 245, G: 240, B: 130, A: 255}
)

// drawAutomap overlays an overhead minimap of the explored level in the top-right
// corner: revealed walls and floor, doors (locked ones tinted by key colour), the
// exit, un-taken items, and the player as a marker with a heading line. It is
// drawn last so nothing occludes it.
func drawAutomap(fb []byte, cfg Config, g *sim.Game) {
	l := g.World.Level
	cell := min(automapMaxW/l.Width, automapMaxH/l.Height)
	if cell < 2 {
		cell = 2
	}
	mapW, mapH := cell*l.Width, cell*l.Height
	ox := cfg.Width - mapW - automapInset
	oy := automapInset

	fillBox(fb, cfg, ox-2, oy-2, mapW+4, mapH+4, mapPanel)

	visited := g.Visited()
	revealed := func(c world.Coord) bool {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if visited[world.Coord{X: c.X + dx, Y: c.Y + dy}] {
					return true
				}
			}
		}
		return false
	}

	for ty := range l.Height {
		for tx := range l.Width {
			c := world.Coord{X: tx, Y: ty}
			if !revealed(c) {
				continue
			}
			if col, ok := tileColor(l, c); ok {
				fillBox(fb, cfg, ox+tx*cell, oy+ty*cell, cell, cell, col)
			}
		}
	}

	// Un-taken items as small dots (keys in their colour, others pale).
	for _, it := range g.Items {
		if it.Taken {
			continue
		}
		c := world.Coord{X: int(it.Pos.X), Y: int(it.Pos.Y)}
		if !revealed(c) {
			continue
		}
		fillBox(fb, cfg, ox+c.X*cell, oy+c.Y*cell, max(cell-1, 1), max(cell-1, 1), itemDotColor(it.Kind))
	}

	drawPlayerMarker(fb, cfg, g, ox, oy, cell)
}

// tileColor returns the minimap colour for a cell, and false if it should be left
// blank (e.g. unexplored solid rock far from any room).
func tileColor(l *world.Level, c world.Coord) (color.RGBA, bool) {
	switch l.At(c.X, c.Y) {
	case world.TileWall:
		return mapWall, true
	case world.TileExit:
		return mapExit, true
	case world.TileDoor:
		if key, locked := l.Locks[c]; locked {
			return keyColor(key), true
		}
		return mapDoor, true
	case world.TileFloor, world.TileSpawn:
		return mapFloor, true
	default:
		return color.RGBA{}, false
	}
}

// drawPlayerMarker draws the player dot and a short heading line on the minimap.
func drawPlayerMarker(fb []byte, cfg Config, g *sim.Game, ox, oy, cell int) {
	mpx := ox + int(g.Player.Pos.X*float64(cell))
	mpy := oy + int(g.Player.Pos.Y*float64(cell))
	fillBox(fb, cfg, mpx-1, mpy-1, 3, 3, mapPlayer)

	dir := g.Player.Dir()
	for t := 0; t < cell*2; t++ {
		setPixel(fb, cfg.Width, mpx+int(dir.X*float64(t)), mpy+int(dir.Y*float64(t)), mapPlayer)
	}
}

// itemDotColor picks a minimap dot colour: keys in their own hue, others pale.
func itemDotColor(k world.ItemKind) color.RGBA {
	if k.IsKey() {
		return keyColor(k)
	}
	return color.RGBA{R: 200, G: 200, B: 180, A: 255}
}

// keyColor maps a keycard kind to its indicator colour.
func keyColor(k world.ItemKind) color.RGBA {
	switch k {
	case world.ItemKeyBlue:
		return color.RGBA{R: 70, G: 110, B: 220, A: 255}
	case world.ItemKeyYellow:
		return color.RGBA{R: 220, G: 200, B: 60, A: 255}
	default:
		return color.RGBA{R: 210, G: 50, B: 50, A: 255}
	}
}

// fillBox fills a w×h rectangle at (x0,y0) with a solid colour, clipped to frame.
func fillBox(fb []byte, cfg Config, x0, y0, w, h int, c color.RGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			setPixel(fb, cfg.Width, x, y, c)
		}
	}
}
