package world

type MarkerKind uint8

const (
	MarkerNone MarkerKind = iota

	MarkerJunction

	MarkerDeadEndDoor

	MarkerDecoyExit
)

func (k MarkerKind) String() string {
	switch k {
	case MarkerJunction:
		return "junction"
	case MarkerDeadEndDoor:
		return "dead_end_door"
	case MarkerDecoyExit:
		return "decoy_exit"
	default:
		return "none"
	}
}

type Marker struct {
	Kind     MarkerKind
	At       Coord
	Branches []Coord
	Optimal  Coord
}
