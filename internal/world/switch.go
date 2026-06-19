package world

// SwitchAction is what pressing a wall switch does.
type SwitchAction uint8

const (
	// SwitchExit ends the level (the level's exit switch).
	SwitchExit SwitchAction = iota
	// SwitchDoor opens the door at the switch's target cell.
	SwitchDoor
)

// Switch records what a TileSwitch wall does when pressed. Target is meaningful
// for SwitchDoor.
type Switch struct {
	Action SwitchAction
	Target Coord
}

// placeExitSwitch turns a wall beside the exit tile into the level's exit switch,
// so finishing a level means reaching it and pressing use rather than walking
// onto a pad. If the exit somehow has no adjacent wall the level keeps a plain
// walk-on exit (the simulation falls back to that when no exit switch exists).
func placeExitSwitch(l *Level) {
	for _, n := range neighbors4(l.Exit) {
		if l.At(n.X, n.Y) != TileWall {
			continue
		}
		l.set(n.X, n.Y, TileSwitch)
		if l.Switches == nil {
			l.Switches = make(map[Coord]Switch)
		}
		l.Switches[n] = Switch{Action: SwitchExit}
		return
	}
}

// SwitchAt returns the switch at (x, y), if any.
func (l *Level) SwitchAt(x, y int) (Switch, bool) {
	s, ok := l.Switches[Coord{X: x, Y: y}]
	return s, ok
}

// HasExitSwitch reports whether the level ends via a switch rather than a walk-on
// exit tile.
func (l *Level) HasExitSwitch() bool {
	for _, s := range l.Switches {
		if s.Action == SwitchExit {
			return true
		}
	}
	return false
}
