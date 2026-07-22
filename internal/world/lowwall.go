package world

const lowWallHeight = 0.4

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

func dividesOpenSpace(l *Level, c Coord) bool {
	openH := l.At(c.X-1, c.Y).Walkable() && l.At(c.X+1, c.Y).Walkable() &&
		l.At(c.X, c.Y-1) == TileWall && l.At(c.X, c.Y+1) == TileWall
	openV := l.At(c.X, c.Y-1).Walkable() && l.At(c.X, c.Y+1).Walkable() &&
		l.At(c.X-1, c.Y) == TileWall && l.At(c.X+1, c.Y) == TileWall
	return openH || openV
}
