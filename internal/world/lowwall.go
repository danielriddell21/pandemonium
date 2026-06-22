package world

// lowWallHeight is how tall a "see-over" low wall stands, in wall units.
const lowWallHeight = 0.4

// placeLowWalls turns a few wall cells that separate two open spaces into low
// walls: still solid (they block movement and sight), but short enough that the
// renderer shows the room beyond over the top, the way DOOM's window ledges do.
// They stay walls, so reachability is unaffected.
func placeLowWalls(l *Level, g *rng) {
	const maxLow = 3
	placed := 0
	for y := 1; y < l.Height-1 && placed < maxLow; y++ {
		for x := 1; x < l.Width-1 && placed < maxLow; x++ {
			if l.At(x, y) != TileWall || l.WallTop[y*l.Width+x] != 0 {
				continue
			}
			// A wall is a clean window divider when it has open floor on exactly
			// one opposing axis (north/south or east/west), so lowering it joins
			// two rooms visually without exposing the surrounding solid mass.
			if !dividesOpenSpace(l, Coord{X: x, Y: y}) {
				continue
			}
			if !g.chance(0.4) {
				continue
			}
			l.WallTop[y*l.Width+x] = lowWallHeight
			placed++
		}
	}
}

// dividesOpenSpace reports whether c is a wall flanked by floor on one axis and
// wall on the other — a thin partition between two spaces.
func dividesOpenSpace(l *Level, c Coord) bool {
	openH := l.At(c.X-1, c.Y).Walkable() && l.At(c.X+1, c.Y).Walkable() &&
		l.At(c.X, c.Y-1) == TileWall && l.At(c.X, c.Y+1) == TileWall
	openV := l.At(c.X, c.Y-1).Walkable() && l.At(c.X, c.Y+1).Walkable() &&
		l.At(c.X-1, c.Y) == TileWall && l.At(c.X+1, c.Y) == TileWall
	return openH || openV
}
