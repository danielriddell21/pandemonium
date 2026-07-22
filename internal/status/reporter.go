package status

import (
	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
	"github.com/danielriddell21/pandemonium/internal/world"
)

const rushThreshold = 0.5

type Reporter struct {
	overlay *hud.Overlay
	src     Source
	profile telemetry.RunProfile

	reach   int
	level   int
	sawKill bool
	sawItem bool
	arrived bool
}

var _ telemetry.Subscriber = (*Reporter)(nil)

type Option func(*Reporter)

func WithReach(level int) Option {
	return func(r *Reporter) {
		if level > 0 {
			r.reach = level
		}
	}
}

func New(overlay *hud.Overlay, src Source, opts ...Option) *Reporter {
	r := &Reporter{overlay: overlay, src: src}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Reporter) OnEvent(e telemetry.PlayerEvent) {
	cue, ok := r.cueFor(e)
	if !ok {
		return
	}
	r.enrich(&cue)
	// At the deepest band the run reaches its terminal beat once, then the voice
	// goes sparse — only the largest moments still draw a line.
	if band(cue.Level) >= deepestBand {
		switch {
		case !r.arrived:
			r.arrived = true
			cue.Arrival = true
		case cue.Kind != CueExit && cue.Kind != CueDeath:
			return
		}
	}
	r.src.Request(cue, r.emit)
}

func (r *Reporter) enrich(c *Cue) {
	p := r.profile
	c.LevelsCleared = p.LevelsCleared
	c.Deaths = p.Deaths
	c.WrongDoorsTotal = p.TotalWrongDoors
	c.ExploreScore = p.ExploreScore
	c.Rushing = p.ExploreScore > 0 && p.ExploreScore < rushThreshold
	if r.reach > c.Level {
		c.Level = r.reach
	}
}

func (r *Reporter) emit(line Line) {
	r.overlay.Post(line.Text, line.Frames, line.Channel)
}

func (r *Reporter) OnPathSummary(telemetry.PathSummary) {}

func (r *Reporter) OnRunProfile(p telemetry.RunProfile) { r.profile = p }

func (r *Reporter) cueFor(e telemetry.PlayerEvent) (Cue, bool) {
	if e.LevelIndex != r.level {
		r.level, r.sawKill, r.sawItem = e.LevelIndex, false, false
	}
	switch e.Kind {
	case sim.ObsExit:
		return Cue{Kind: CueExit, Level: e.LevelIndex}, true
	case sim.ObsDeath:
		return Cue{Kind: CueDeath, Level: e.LevelIndex}, true
	case sim.ObsSecret:
		return Cue{Kind: CueSecret, Level: e.LevelIndex}, true
	case sim.ObsKill:
		if r.sawKill {
			break
		}
		r.sawKill = true
		return Cue{Kind: CueKill, Level: e.LevelIndex}, true
	case sim.ObsItem:
		if r.sawItem {
			break
		}
		r.sawItem = true
		return Cue{Kind: CueItem, Level: e.LevelIndex}, true
	case sim.ObsDoor:
		if e.Marker != nil && e.Marker.WrongDoor {
			return Cue{Kind: CueWrongDoor, Level: e.LevelIndex}, true
		}
	case sim.ObsMarker:
		return r.markerCue(e)
	}
	return Cue{}, false
}

func (r *Reporter) markerCue(e telemetry.PlayerEvent) (Cue, bool) {
	if e.Marker == nil {
		return Cue{}, false
	}
	switch e.Marker.KindEnum {
	case world.MarkerDecoyExit:
		return Cue{Kind: CueDecoy, Level: e.LevelIndex}, true
	case world.MarkerJunction:
		optimal := e.Marker.TakenX == e.Marker.OptimalX && e.Marker.TakenY == e.Marker.OptimalY
		return Cue{Kind: CueFork, Level: e.LevelIndex, OptimalChoice: optimal}, true
	}
	return Cue{}, false
}
