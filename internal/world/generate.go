package world

import (
	"errors"
	"fmt"
)

const minDimension = 16

var ErrUnreachable = errors.New("world: exhausted attempts producing a connected level")

type Config struct {
	Width, Height int

	Seed int64

	MaxAttempts int

	Arena bool
}

func (c Config) normalized() Config {
	if c.Width < minDimension {
		c.Width = minDimension
	}
	if c.Height < minDimension {
		c.Height = minDimension
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 32
	}
	return c
}

func Generate(cfg Config) (*Level, error) {
	cfg = cfg.normalized()
	for attempt := range cfg.MaxAttempts {
		// Derive a per-attempt seed deterministically from the base seed.
		sub := cfg.Seed + int64(attempt)*0x100000001b3
		l := generateOnce(cfg.Width, cfg.Height, sub, cfg.Arena)
		l.Seed = cfg.Seed
		// The exit must be reachable once doors are open, and every keycard must
		// be obtainable without first crossing the door it unlocks.
		if reachable(l, l.Spawn, l.Exit, blocksWalls(l)) && keysReachable(l) {
			return l, nil
		}
	}
	return nil, fmt.Errorf("%w: %dx%d seed=%d", ErrUnreachable, cfg.Width, cfg.Height, cfg.Seed)
}

func generateOnce(width, height int, seed int64, arena bool) *Level {
	if arena {
		return generateArena(width, height, seed)
	}
	l := newLevel(width, height, seed)
	g := newRNG(seed)

	root := &bspNode{bounds: rect{x: 1, y: 1, w: width - 2, h: height - 2}}
	g.split(root, 0)
	g.carveRooms(root, l)
	g.connect(root, l)
	g.carveStubs(l)

	rooms := collectRooms(root)
	placeSpawnAndExit(l, rooms)
	placeExitSwitch(l)
	annotate(l)
	placeKeyGate(l, g)
	placeItems(l, g)
	placeBarrels(l, g)
	assignHeights(l, g, rooms)
	placeHazards(l, g)
	assignLight(l, g, rooms)
	assignThemes(l, g, rooms)
	placeLowWalls(l, g)
	assignSky(l, g, rooms) // last: purely additive, leaves earlier stages untouched
	return l
}

func placeSpawnAndExit(l *Level, rooms []rect) {
	if len(rooms) == 0 {
		return
	}
	spawn := rooms[0].center()
	dist := distanceField(l, spawn)
	exit, best := spawn, 0
	for i, d := range dist {
		if d > best {
			best = d
			exit = Coord{X: i % l.Width, Y: i / l.Width}
		}
	}
	l.Spawn = spawn
	l.set(spawn.X, spawn.Y, TileSpawn)
	if exit != spawn {
		l.Exit = exit
		l.set(exit.X, exit.Y, TileExit)
	}
}
