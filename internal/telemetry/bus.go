package telemetry

import (
	"strconv"
	"time"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

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

var _ sim.Observer = (*Bus)(nil)

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

func (b *Bus) BeginLevel(seed int64, index int) {
	if b.pathOpen {
		b.finishPath()
	}
	b.path = PathSummary{LevelSeed: seed, LevelIndex: index}
	b.visited = make(map[[2]int]bool)
	b.levelStart = b.now()
	b.pathOpen = true
}

func (b *Bus) Observe(o sim.Observation) {
	ev := PlayerEvent{
		Tick:        o.Tick,
		TimestampMS: b.now().Sub(b.start).Milliseconds(),
		Type:        o.Kind.String(),
		Kind:        o.Kind,
		LevelSeed:   b.path.LevelSeed,
		LevelIndex:  b.path.LevelIndex,
		Cell:        [2]int{o.At.X, o.At.Y},
		Marker:      markerFor(o),
	}

	if o.Kind == sim.ObsDeath {
		b.profile.Deaths++
	}
	if b.pathOpen {
		b.recordPath(o)
	}

	b.dispatchEvent(ev)

	if o.Kind == sim.ObsExit && b.pathOpen {
		b.finishPath()
	}
}

func markerFor(o sim.Observation) *MarkerInfo {
	switch o.Kind {
	case sim.ObsMarker:
		return markerInfo(o)
	case sim.ObsDoor:
		return &MarkerInfo{Kind: "door", X: o.At.X, Y: o.At.Y, WrongDoor: o.WrongDoor}
	}
	return nil
}

func (b *Bus) recordPath(o sim.Observation) {
	switch o.Kind {
	case sim.ObsMove:
		b.path.Steps++
		key := [2]int{o.At.X, o.At.Y}
		if b.visited[key] {
			b.path.BacktrackCount++
		} else {
			b.visited[key] = true
			b.path.TilesVisited++
		}
	case sim.ObsMarker:
		if o.Marker == world.MarkerJunction {
			b.path.JunctionsSeen++
			if o.Taken == o.Optimal {
				b.path.OptimalChoices++
			}
		}
	case sim.ObsDoor:
		b.path.DoorsOpened++
		if o.WrongDoor {
			b.path.WrongDoors++
		}
	case sim.ObsKill:
		b.path.Kills++
	case sim.ObsItem:
		b.path.ItemsTaken++
	case sim.ObsSecret:
		b.path.SecretsFound++
	case sim.ObsExit:
		b.path.Completed = true
	}
}

func (b *Bus) Profile() RunProfile { return b.profile }

func (b *Bus) finishPath() {
	b.path.TimeSpentMS = b.now().Sub(b.levelStart).Milliseconds()
	if b.path.Completed {
		b.profile.LevelsCleared++
	}
	b.profile.TotalSteps += b.path.Steps
	b.profile.TotalDoorsOpened += b.path.DoorsOpened
	b.profile.TotalWrongDoors += b.path.WrongDoors
	b.profile.TotalKills += b.path.Kills
	b.profile.TotalItems += b.path.ItemsTaken
	b.profile.TotalSecrets += b.path.SecretsFound
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

func markerInfo(o sim.Observation) *MarkerInfo {
	mi := &MarkerInfo{
		Kind:      o.Marker.String(),
		KindEnum:  o.Marker,
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
