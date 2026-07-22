package world

const arenaEvery = 5

func IsArenaLevel(levelNumber int) bool {
	return levelNumber > 0 && levelNumber%arenaEvery == 0
}

func generateArena(width, height int, seed int64) *Level {
	l := newLevel(width, height, seed)
	g := newRNG(seed)

	// Carve one big rectangle inside the border and open it to the sky.
	x0, y0, x1, y1 := 2, 2, width-3, height-3
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			i := y*width + x
			l.Tiles[i] = TileFloor
			l.Sky[i] = true
			l.Light[i] = 1
			l.CeilH[i] = skyHeadroom
		}
	}

	midY := (y0 + y1) / 2
	l.Spawn = Coord{X: x0, Y: midY}
	l.set(l.Spawn.X, l.Spawn.Y, TileSpawn)
	l.Exit = Coord{X: x1, Y: midY}
	l.set(l.Exit.X, l.Exit.Y, TileExit)
	placeExitSwitch(l)

	placeArenaCache(l)
	placeItems(l, g)   // plus the usual scattered consumables
	placeBarrels(l, g) // and barrels for crossfire
	return l
}

func placeArenaCache(l *Level) {
	cx, cy := l.Width/2, l.Height/2
	cache := []struct {
		kind ItemKind
		dx   int
	}{
		{ItemMega, 0},
		{ItemArmor, -1},
		{ItemRockets, 1},
	}
	for _, c := range cache {
		at := Coord{X: cx + c.dx, Y: cy}
		if l.At(at.X, at.Y) == TileFloor {
			l.Items = append(l.Items, Item{Kind: c.kind, At: at})
		}
	}
}
