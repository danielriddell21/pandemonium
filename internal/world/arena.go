package world

// arenaEvery is how often a set-piece arena replaces the usual maze: every Nth
// level (1-based) opens into one large, sky-lit room with a heavier fight.
const arenaEvery = 5

// IsArenaLevel reports whether the given 1-based level number is a set-piece
// arena rather than an ordinary generated maze.
func IsArenaLevel(levelNumber int) bool {
	return levelNumber > 0 && levelNumber%arenaEvery == 0
}

// generateArena builds a single large, open, sky-lit room: the spawn at one end,
// the exit switch at the other, a weapon-and-armour cache at its heart and a
// scatter of barrels. The encounter is filled by the simulation, which scales
// the demon count with the (large) floor area, so an arena reads as a heavier,
// open fight than the usual corridors.
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

// placeArenaCache lays a guaranteed reward at the heart of the arena — a
// megasphere flanked by armour and rockets — so the open fight comes with the
// firepower to match.
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
