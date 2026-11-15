package gui

import (
	"fmt"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
	"github.com/danielriddell21/pandemonium/internal/world"
)

type session struct {
	baseSeed      int64
	width, height int
	bus           *telemetry.Bus
	audio         sim.Observer
	level         int
	skill         sim.Skill
	attractSeed   int64
	current       *sim.Game
}

const attractSeedBase = 1 << 40

func newSession(baseSeed int64, width, height int, bus *telemetry.Bus, audio *Audio) *session {
	s := &session{baseSeed: baseSeed, width: width, height: height, bus: bus, skill: sim.SkillNormal}
	if audio != nil { // keep the observer slot a true nil when there is no engine
		s.audio = audio
	}
	return s
}

func (s *session) setSkill(skill int) {
	if skill < int(sim.SkillEasy) || skill > int(sim.SkillNightmare) {
		skill = int(sim.SkillNormal)
	}
	s.skill = sim.Skill(skill)
}

func (s *session) start() *sim.Game {
	return s.build()
}

func (s *session) next() *sim.Game {
	s.level++
	// What the player walked out with is what they walk in with.
	return s.build(sim.WithLoadout(s.current.Loadout()))
}

func (s *session) build(carry ...sim.Option) *sim.Game {
	arena := world.IsArenaLevel(s.level + 1) // 1-based: every fifth level is a set-piece
	lvl, seed := generateLevel(world.Config{Width: s.width, Height: s.height, Seed: s.baseSeed + int64(s.level), Arena: arena})

	s.bus.BeginLevel(seed, s.level)
	fmt.Printf("level %d  seed %d  (%dx%d)\n", s.level+1, seed, s.width, s.height)
	// Telemetry always observes; the sound engine observes too when present.
	opts := append([]sim.Option{
		sim.WithSkill(s.skill),
		sim.WithDepth(s.level),
		sim.WithObserver(sim.Fanout(s.bus, s.audio)),
	}, carry...)
	s.current = sim.New(lvl, opts...)
	return s.current
}

func (s *session) attract() *sim.Game {
	s.attractSeed++
	lvl, _ := generateLevel(world.Config{Width: s.width, Height: s.height, Seed: attractSeedBase + s.attractSeed})
	return sim.New(lvl)
}

const maxLevelSeeds = 64

func generateLevel(cfg world.Config) (*world.Level, int64) {
	for attempt := 0; ; attempt++ {
		l, err := world.Generate(cfg)
		if err == nil {
			return l, cfg.Seed
		}
		if attempt >= maxLevelSeeds {
			// Every seed failed: the dimensions cannot hold a connected level.
			panic(fmt.Errorf("pandemonium: no reachable %dx%d level after %d seeds: %w",
				cfg.Width, cfg.Height, attempt+1, err))
		}
		cfg.Seed++
	}
}
