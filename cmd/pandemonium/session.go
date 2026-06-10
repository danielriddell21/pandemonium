package main

import (
	"fmt"

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
	level         int
}

// newSession starts a run at the given base seed and level dimensions.
func newSession(baseSeed int64, width, height int, bus *telemetry.Bus) *session {
	return &session{baseSeed: baseSeed, width: width, height: height, bus: bus}
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
	return sim.New(lvl, sim.WithObserver(s.bus))
}
