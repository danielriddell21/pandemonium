package sim

import "github.com/danielriddell21/pandemonium/internal/world"

type ObservationKind uint8

const (
	ObsMove ObservationKind = iota

	ObsMarker

	ObsDoor

	ObsExit

	ObsDeath

	ObsItem

	ObsSecret

	ObsKill
)

func (k ObservationKind) String() string {
	switch k {
	case ObsMove:
		return "move"
	case ObsMarker:
		return "marker"
	case ObsDoor:
		return "door"
	case ObsExit:
		return "exit"
	case ObsDeath:
		return "death"
	case ObsItem:
		return "item"
	case ObsSecret:
		return "secret"
	case ObsKill:
		return "kill"
	default:
		return "unknown"
	}
}

type Observation struct {
	Tick      uint64
	Kind      ObservationKind
	At        world.Coord
	Marker    world.MarkerKind
	Taken     world.Coord
	Optimal   world.Coord
	Ignored   []world.Coord
	WrongDoor bool
}

type Observer interface {
	Observe(Observation)
}

type nopObserver struct{}

func (nopObserver) Observe(Observation) {}

func Fanout(obs ...Observer) Observer {
	return fanout(obs)
}

type fanout []Observer

func (f fanout) Observe(o Observation) {
	for _, ob := range f {
		if ob != nil {
			ob.Observe(o)
		}
	}
}

type Option func(*Game)

func WithObserver(o Observer) Option {
	return func(g *Game) {
		if o != nil {
			g.observer = o
		}
	}
}
