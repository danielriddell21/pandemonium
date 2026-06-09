package world

// MarkerKind classifies a tagged structure in a generated level. These tags
// describe the shape of the map for layout analysis: where it branches, where it
// dead-ends behind a door, and where a tile resembles the exit without being it.
type MarkerKind uint8

const (
	// MarkerNone is the zero value and tags nothing.
	MarkerNone MarkerKind = iota
	// MarkerJunction tags a branch point where the layout forks.
	MarkerJunction
	// MarkerDeadEndDoor tags a door whose only destination is a dead end.
	MarkerDeadEndDoor
	// MarkerDecoyExit tags a tile that resembles the exit but is not it.
	MarkerDecoyExit
)

// String returns a stable label for the kind, suitable for logs and analytics.
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

// Marker is layout metadata attached to a level: a structurally interesting cell
// and, where relevant, the branches available from it and which branch lies on
// the shortest route toward the exit.
type Marker struct {
	Kind     MarkerKind
	At       Coord
	Branches []Coord // candidate directions out of At (junctions, doors)
	Optimal  Coord   // the branch closest to the exit, when meaningful
}
