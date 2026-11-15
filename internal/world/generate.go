package world

import (
	"errors"
	"fmt"

	"github.com/danielriddell21/crucible/level"
	"github.com/danielriddell21/crucible/worldgen"
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
	if cfg.Arena {
		return generateArena(cfg.Width, cfg.Height, cfg.Seed), nil
	}
	deck := &Level{}
	furnish := func(l *level.Level, g *worldgen.RNG, rooms []rect) {
		deck.reset(l)
		placeExitSwitch(deck)
		annotate(deck)
		placeKeyGate(deck, g)
		placeItems(deck, g)
		placeBarrels(deck, g)
		placeAlarmPanel(deck)
		level.AssignHeights(l, g, rooms, level.HeightsConfig{})
		placeLift(deck, g)
		placeHazards(deck, g)
		assignLight(deck, g, rooms)
		level.AssignThemes(l, g, rooms, NumThemes)
		level.PlaceLowWalls(l, g, level.LowWallConfig{})
		level.AssignSky(l, g, rooms, level.SkyConfig{}) // last: purely additive
	}
	validate := func(*level.Level) bool {
		return reachable(deck, deck.Spawn, deck.Exit, blocksWalls(deck)) && keysReachable(deck)
	}
	base, _, err := level.Generate(level.GenerateConfig{
		Width:  cfg.Width,
		Height: cfg.Height,
		Seed:   cfg.Seed,
	}, []level.Pass{furnish}, validate)
	if err != nil {
		return nil, fmt.Errorf("%w: %dx%d seed=%d", ErrUnreachable, cfg.Width, cfg.Height, cfg.Seed)
	}
	deck.Level = base
	return deck, nil
}

func (l *Level) reset(base *level.Level) {
	*l = Level{
		Level:    base,
		Locks:    map[Coord]ItemKind{},
		Hazard:   map[Coord]HazardCell{},
		Switches: map[Coord]Switch{},
	}
}

func placeLift(l *Level, g *worldgen.RNG) {
	taken := make(map[Coord]bool, len(l.Items)+len(l.Secrets))
	for _, it := range l.Items {
		taken[it.At] = true
	}
	for _, s := range l.Secrets {
		taken[s] = true
	}
	if ledge, ok := level.PlaceLiftLedge(l.Level, g, level.LiftConfig{}, func(c Coord) bool { return taken[c] }); ok {
		l.Items = append(l.Items, Item{Kind: secretReward(g), At: ledge})
	}
}
