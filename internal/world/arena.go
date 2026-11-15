package world

import "github.com/danielriddell21/crucible/worldgen"

const ArenaEvery = 5

const arenaSkyHeadroom = 3.0

func IsArenaLevel(levelNumber int) bool {
	return levelNumber > 0 && levelNumber%ArenaEvery == 0
}

func generateArena(width, height int, seed int64) *Level {
	l := newLevel(width, height, seed)
	g := worldgen.NewRNG(seed)

	// Carve one big rectangle inside the border and open it to the sky.
	x0, y0, x1, y1 := 2, 2, width-3, height-3
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			i := y*width + x
			l.Tiles[i] = TileFloor
			l.Sky[i] = true
			l.Light[i] = 1
			l.CeilH[i] = arenaSkyHeadroom
		}
	}

	midY := (y0 + y1) / 2
	l.Spawn = Coord{X: x0, Y: midY}
	l.Set(l.Spawn.X, l.Spawn.Y, TileSpawn)
	l.Exit = Coord{X: x1, Y: midY}
	l.Set(l.Exit.X, l.Exit.Y, TileExit)
	placeExitSwitch(l)

	placeArenaPillars(l, x0, y0, x1, y1)
	l.Boss = true
	l.BossAt = Coord{X: x1 - (x1-x0)/4, Y: midY}

	placeArenaCache(l)
	placeItems(l, g)   // plus the usual scattered consumables
	placeBarrels(l, g) // and barrels for crossfire
	return l
}

func placeArenaPillars(l *Level, x0, y0, x1, y1 int) {
	// Four blocks of cover off the centre line: somewhere to break a volley's
	// line of fire, and something solid to bait a charge into.
	midY := (y0 + y1) / 2
	dx, dy := (x1-x0)/3, (y1-y0)/4
	for _, c := range [4]Coord{
		{X: x0 + dx, Y: midY - dy - 1},
		{X: x1 - dx - 1, Y: midY - dy - 1},
		{X: x0 + dx, Y: midY + dy},
		{X: x1 - dx - 1, Y: midY + dy},
	} {
		for y := c.Y; y <= c.Y+1; y++ {
			for x := c.X; x <= c.X+1; x++ {
				i := y*l.W + x
				l.Tiles[i] = TileWall
				l.Sky[i] = false
			}
		}
	}
}

func placeArenaCache(l *Level) {
	cx, cy := l.W/2, l.H/2
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
