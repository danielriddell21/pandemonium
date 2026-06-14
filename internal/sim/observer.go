package sim

import "github.com/danielriddell21/pandemonium/internal/world"

// ObservationKind categorises a single thing the simulation noticed the player
// do this tick. The simulation emits these as the player moves through a level.
type ObservationKind uint8

const (
	// ObsMove reports the player entering a new tile.
	ObsMove ObservationKind = iota
	// ObsMarker reports the player crossing a tagged structure (see Marker).
	ObsMarker
	// ObsDoor reports the player opening a door.
	ObsDoor
	// ObsExit reports the player reaching the exit.
	ObsExit
	// ObsDeath reports the player dying.
	ObsDeath
	// ObsItem reports the player collecting an item.
	ObsItem
	// ObsSecret reports the player discovering a secret area.
	ObsSecret
)

// String returns a stable label for the kind.
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
	default:
		return "unknown"
	}
}

// Observation is a single structured record of player activity, expressed purely
// in world terms. For markers it records which branch the player took and which
// they passed over.
type Observation struct {
	Tick      uint64
	Kind      ObservationKind
	At        world.Coord
	Marker    world.MarkerKind // MarkerNone unless Kind == ObsMarker
	Taken     world.Coord      // branch entered (junctions)
	Optimal   world.Coord      // branch nearest the exit (junctions)
	Ignored   []world.Coord    // branches not entered (junctions)
	WrongDoor bool             // door opened led only to a dead end
}

// Observer consumes observations emitted by the simulation. The simulation holds
// exactly one and always calls it; the default does nothing.
type Observer interface {
	Observe(Observation)
}

// nopObserver is the default sink: it discards every observation.
type nopObserver struct{}

func (nopObserver) Observe(Observation) {}

// Option configures a Game at construction.
type Option func(*Game)

// WithObserver attaches an observer to receive the simulation's observations.
func WithObserver(o Observer) Option {
	return func(g *Game) {
		if o != nil {
			g.observer = o
		}
	}
}
