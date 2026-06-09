package telemetry

import (
	"strconv"
	"time"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// Bus turns the simulation's observations into telemetry: it adapts each
// observation into a PlayerEvent, maintains a rolling per-level PathSummary and a
// cumulative RunProfile, and fans the results out to its subscribers. It
// satisfies sim.Observer, so it can be attached to a game directly.
type Bus struct {
	subs    []Subscriber
	profile RunProfile
	path    PathSummary
	visited map[[2]int]bool

	start      time.Time
	levelStart time.Time
	pathOpen   bool

	now func() time.Time
}

// Compile-time check that Bus can be used as the simulation's observer.
var _ sim.Observer = (*Bus)(nil)

// NewBus creates a bus that dispatches to the given subscribers.
func NewBus(subs ...Subscriber) *Bus {
	now := time.Now
	start := now()
	return &Bus{
		subs:  subs,
		start: start,
		now:   now,
		profile: RunProfile{
			SessionID: strconv.FormatInt(start.UnixNano(), 36),
			StartedMS: start.UnixMilli(),
		},
	}
}

// BeginLevel starts tracking a fresh level. Any in-progress level is finalised
// first, so an abandoned level is still recorded.
func (b *Bus) BeginLevel(seed int64, index int) {
	if b.pathOpen {
		b.finishPath()
	}
	b.path = PathSummary{LevelSeed: seed, LevelIndex: index}
	b.visited = make(map[[2]int]bool)
	b.levelStart = b.now()
	b.pathOpen = true
}

// Observe consumes one observation from the simulation.
func (b *Bus) Observe(o sim.Observation) {
	ev := PlayerEvent{
		Tick:        o.Tick,
		TimestampMS: b.now().Sub(b.start).Milliseconds(),
		Type:        o.Kind.String(),
		LevelSeed:   b.path.LevelSeed,
		LevelIndex:  b.path.LevelIndex,
		Cell:        [2]int{o.At.X, o.At.Y},
	}

	switch o.Kind {
	case sim.ObsMove:
		if b.pathOpen {
			b.path.Steps++
			key := [2]int{o.At.X, o.At.Y}
			if b.visited[key] {
				b.path.BacktrackCount++
			} else {
				b.visited[key] = true
				b.path.TilesVisited++
			}
		}
	case sim.ObsMarker:
		ev.Marker = markerInfo(o)
		if b.pathOpen && o.Marker == world.MarkerJunction {
			b.path.JunctionsSeen++
			if o.Taken == o.Optimal {
				b.path.OptimalChoices++
			}
		}
	case sim.ObsDoor:
		ev.Marker = &MarkerInfo{Kind: "door", X: o.At.X, Y: o.At.Y, WrongDoor: o.WrongDoor}
		if b.pathOpen {
			b.path.DoorsOpened++
			if o.WrongDoor {
				b.path.WrongDoors++
			}
		}
	case sim.ObsDeath:
		b.profile.Deaths++
	case sim.ObsExit:
		if b.pathOpen {
			b.path.Completed = true
		}
	}

	b.dispatchEvent(ev)

	if o.Kind == sim.ObsExit && b.pathOpen {
		b.finishPath()
	}
}

// Profile returns a snapshot of the current run profile.
func (b *Bus) Profile() RunProfile { return b.profile }

// finishPath closes the current level's summary, folds it into the run profile,
// and notifies subscribers.
func (b *Bus) finishPath() {
	b.path.TimeSpentMS = b.now().Sub(b.levelStart).Milliseconds()
	if b.path.Completed {
		b.profile.LevelsCleared++
	}
	b.profile.TotalSteps += b.path.Steps
	b.profile.TotalDoorsOpened += b.path.DoorsOpened
	b.profile.TotalWrongDoors += b.path.WrongDoors
	b.profile.ChoicesPerLevel = append(b.profile.ChoicesPerLevel, b.path.JunctionsSeen)
	b.profile.Paths = append(b.profile.Paths, b.path)
	b.profile.ExploreScore = exploreScore(b.profile)
	b.pathOpen = false

	b.dispatchPath(b.path)
	b.dispatchProfile(b.profile)
}

func (b *Bus) dispatchEvent(ev PlayerEvent) {
	for _, s := range b.subs {
		s.OnEvent(ev)
	}
}

func (b *Bus) dispatchPath(p PathSummary) {
	for _, s := range b.subs {
		s.OnPathSummary(p)
	}
}

func (b *Bus) dispatchProfile(p RunProfile) {
	for _, s := range b.subs {
		s.OnRunProfile(p)
	}
}

// markerInfo builds the marker payload for a marker observation.
func markerInfo(o sim.Observation) *MarkerInfo {
	mi := &MarkerInfo{
		Kind:      o.Marker.String(),
		X:         o.At.X,
		Y:         o.At.Y,
		WrongDoor: o.WrongDoor,
	}
	if o.Marker == world.MarkerJunction {
		mi.TakenX, mi.TakenY = o.Taken.X, o.Taken.Y
		mi.OptimalX, mi.OptimalY = o.Optimal.X, o.Optimal.Y
		for _, ig := range o.Ignored {
			mi.Ignored = append(mi.Ignored, [2]int{ig.X, ig.Y})
		}
	}
	return mi
}

// exploreScore is a heuristic in [0,1]: the share of moves that reached a new
// tile rather than retreading old ground. High means a thorough explorer; low
// means a rusher who beelines and doubles back little.
func exploreScore(p RunProfile) float64 {
	visited := 0
	for _, s := range p.Paths {
		visited += s.TilesVisited
	}
	if p.TotalSteps == 0 {
		return 0
	}
	return float64(visited) / float64(p.TotalSteps)
}
