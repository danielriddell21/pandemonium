package sim

import "github.com/danielriddell21/pandemonium/internal/world"

// tracker watches the player's progress through a level and decides when the
// simulation should emit an observation: on entering a new tile, on crossing a
// tagged structure, and on reaching the exit (once).
type tracker struct {
	markers     map[world.Coord]world.Marker
	lastCell    world.Coord
	started     bool
	exitEmitted bool
}

func newTracker(l *world.Level) tracker {
	m := make(map[world.Coord]world.Marker, len(l.Markers))
	for _, mk := range l.Markers {
		m[mk.At] = mk
	}
	return tracker{markers: m}
}

// markerAt returns the marker tagged at c, if any.
func (t *tracker) markerAt(c world.Coord) (world.Marker, bool) {
	mk, ok := t.markers[c]
	return mk, ok
}

// splitBranches decides which branch of a junction the player took — the one
// best aligned with their facing — and returns the rest as ignored.
func splitBranches(mk world.Marker, dir Vec2, cell world.Coord) (taken world.Coord, ignored []world.Coord) {
	taken = mk.Optimal
	best := -2.0
	for _, b := range mk.Branches {
		vx, vy := float64(b.X-cell.X), float64(b.Y-cell.Y)
		if dot := vx*dir.X + vy*dir.Y; dot > best {
			best, taken = dot, b
		}
	}
	for _, b := range mk.Branches {
		if b != taken {
			ignored = append(ignored, b)
		}
	}
	return taken, ignored
}
