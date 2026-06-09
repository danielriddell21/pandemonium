package world

import (
	"errors"
	"fmt"
)

// Minimum grid size that can hold a sensible partition.
const minDimension = 16

// ErrUnreachable is returned when generation cannot produce a level whose exit
// is reachable from its spawn within the attempt budget.
var ErrUnreachable = errors.New("world: exhausted attempts producing a connected level")

// Config controls level generation.
type Config struct {
	// Width and Height are the grid dimensions in tiles.
	Width, Height int
	// Seed makes generation deterministic: the same Config yields the same Level.
	Seed int64
	// MaxAttempts bounds how many times generation retries when a candidate
	// level fails the reachability guarantee. Zero selects a sensible default.
	MaxAttempts int
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

// Generate produces a level from cfg. It repeatedly builds candidates until one
// has its exit reachable from its spawn (the flood-fill guarantee), then returns
// it. Each attempt derives a distinct but deterministic sub-seed, so a given
// Config always yields an identical Level.
func Generate(cfg Config) (*Level, error) {
	cfg = cfg.normalized()
	for attempt := range cfg.MaxAttempts {
		// Derive a per-attempt seed deterministically from the base seed.
		sub := cfg.Seed + int64(attempt)*0x100000001b3
		l := generateOnce(cfg.Width, cfg.Height, sub)
		l.Seed = cfg.Seed
		if Reachable(l, l.Spawn, l.Exit) {
			return l, nil
		}
	}
	return nil, fmt.Errorf("%w: %dx%d seed=%d", ErrUnreachable, cfg.Width, cfg.Height, cfg.Seed)
}

// generateOnce builds a single candidate level: partition, carve rooms, connect
// them, then place spawn and exit in two far-apart rooms.
func generateOnce(width, height int, seed int64) *Level {
	l := newLevel(width, height, seed)
	g := newRNG(seed)

	root := &bspNode{bounds: rect{x: 1, y: 1, w: width - 2, h: height - 2}}
	g.split(root, 0)
	g.carveRooms(root, l)
	g.connect(root, l)

	placeSpawnAndExit(l, collectRooms(root))
	annotate(l)
	return l
}

// placeSpawnAndExit marks the first room's centre as spawn and the centre of the
// room farthest from it (by Manhattan distance) as exit, maximising the journey.
func placeSpawnAndExit(l *Level, rooms []rect) {
	if len(rooms) == 0 {
		return
	}
	spawn := rooms[0].center()
	exit := spawn
	best := -1
	for _, r := range rooms[1:] {
		c := r.center()
		if d := manhattan(spawn, c); d > best {
			best, exit = d, c
		}
	}
	l.Spawn = spawn
	l.Exit = exit
	l.set(spawn.X, spawn.Y, TileSpawn)
	l.set(exit.X, exit.Y, TileExit)
}

func manhattan(a, b Coord) int {
	dx, dy := a.X-b.X, a.Y-b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}
