package render

import (
	"image/color"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

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

func drawAutomap(fb []byte, cfg Config, g *sim.Game) {
	l := g.World.Level
	cell := min(automapMaxW/l.W, automapMaxH/l.H)
	if cell < 2 {
		cell = 2
	}
	mapW, mapH := cell*l.W, cell*l.H
	ox := cfg.Width - mapW - automapInset
	oy := automapInset

	fillBox(fb, cfg, ox-2, oy-2, mapW+4, mapH+4, mapPanel)

	visited := g.Visited()
	drawMapTiles(fb, cfg, l, visited, ox, oy, cell)
	drawMapItems(fb, cfg, g, visited, ox, oy, cell)
	drawPlayerMarker(fb, cfg, g, ox, oy, cell)
}

func revealedCell(visited map[world.Coord]bool, c world.Coord) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if visited[world.Coord{X: c.X + dx, Y: c.Y + dy}] {
				return true
			}
		}
	}
	return false
}

func drawMapTiles(fb []byte, cfg Config, l *world.Level, visited map[world.Coord]bool, ox, oy, cell int) {
	for ty := range l.H {
		for tx := range l.W {
			c := world.Coord{X: tx, Y: ty}
			if !revealedCell(visited, c) {
				continue
			}
			if col, ok := tileColor(l, c); ok {
				fillBox(fb, cfg, ox+tx*cell, oy+ty*cell, cell, cell, col)
			}
		}
	}
}

func drawMapItems(fb []byte, cfg Config, g *sim.Game, visited map[world.Coord]bool, ox, oy, cell int) {
	for _, it := range g.Items {
		if it.Taken {
			continue
		}
		c := world.Coord{X: int(it.Pos.X), Y: int(it.Pos.Y)}
		if !revealedCell(visited, c) {
			continue
		}
		fillBox(fb, cfg, ox+c.X*cell, oy+c.Y*cell, max(cell-1, 1), max(cell-1, 1), itemDotColor(it.Kind))
	}
}

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

func drawPlayerMarker(fb []byte, cfg Config, g *sim.Game, ox, oy, cell int) {
	mpx := ox + int(g.Player.Pos.X*float64(cell))
	mpy := oy + int(g.Player.Pos.Y*float64(cell))
	fillBox(fb, cfg, mpx-1, mpy-1, 3, 3, mapPlayer)

	dir := g.Player.Dir()
	for t := 0; t < cell*2; t++ {
		setPixel(fb, cfg.Width, mpx+int(dir.X*float64(t)), mpy+int(dir.Y*float64(t)), mapPlayer)
	}
}

func itemDotColor(k world.ItemKind) color.RGBA {
	if k.IsKey() {
		return keyColor(k)
	}
	return color.RGBA{R: 200, G: 200, B: 180, A: 255}
}

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

func fillBox(fb []byte, cfg Config, x0, y0, w, h int, c color.RGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			setPixel(fb, cfg.Width, x, y, c)
		}
	}
}
