package main

import (
	"fmt"

	"github.com/danielriddell21/pandemonium/internal/app"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// session ties one play-through together: it generates successive levels from a
// base seed, wires each level's simulation to the telemetry bus, and reports the
// seed and level number as it goes.
type session struct {
	baseSeed      int64
	width, height int
	bus           *telemetry.Bus
	audio         sim.Observer // optional sound engine; nil when audio is off
	level         int
	attractSeed   int64 // advances each attract level, offset far from the run's seeds
}

// attractSeedBase keeps the title-screen demo worlds well clear of any plausible
// run seed so the showcase never mirrors the player's actual first level.
const attractSeedBase = 1 << 40

// newSession starts a run at the given base seed and level dimensions. audio may
// be nil, in which case the game runs silent.
func newSession(baseSeed int64, width, height int, bus *telemetry.Bus, audio *app.Audio) *session {
	s := &session{baseSeed: baseSeed, width: width, height: height, bus: bus}
	if audio != nil { // keep the observer slot a true nil when there is no engine
		s.audio = audio
	}
	return s
}

// start builds the first level's simulation.
func (s *session) start() *sim.Game {
	return s.build()
}

// next advances to the following level and builds its simulation. It is handed
// to the app, which calls it whenever the player reaches an exit.
func (s *session) next() *sim.Game {
	s.level++
	return s.build()
}

// build generates the current level (retrying on the rare unreachable case) and
// returns a simulation observed by the telemetry bus.
func (s *session) build() *sim.Game {
	seed := s.baseSeed + int64(s.level)
	var lvl *world.Level
	for {
		l, err := world.Generate(world.Config{Width: s.width, Height: s.height, Seed: seed})
		if err == nil {
			lvl = l
			break
		}
		seed++
	}

	s.bus.BeginLevel(seed, s.level)
	fmt.Printf("level %d  seed %d  (%dx%d)\n", s.level+1, seed, s.width, s.height)
	// Telemetry always observes; the sound engine observes too when present.
	return sim.New(lvl, sim.WithObserver(sim.Fanout(s.bus, s.audio)))
}

// attract builds a throwaway level for the title screen's demo loop: a fresh
// world the bot can play, drawn from seeds well clear of the run's own, and
// observed by nothing (the attract loop must not feed telemetry or sound).
func (s *session) attract() *sim.Game {
	s.attractSeed++
	seed := attractSeedBase + s.attractSeed
	var lvl *world.Level
	for {
		l, err := world.Generate(world.Config{Width: s.width, Height: s.height, Seed: seed})
		if err == nil {
			lvl = l
			break
		}
		seed++
	}
	return sim.New(lvl)
}
