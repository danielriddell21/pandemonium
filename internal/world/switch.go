package world

type SwitchAction uint8

const (
	SwitchExit SwitchAction = iota

	SwitchDoor
)

type Switch struct {
	Action SwitchAction
	Target Coord
}

func placeExitSwitch(l *Level) {
	for _, n := range neighbors4(l.Exit) {
		if l.At(n.X, n.Y) != TileWall {
			continue
		}
		l.Set(n.X, n.Y, TileSwitch)
		if l.Switches == nil {
			l.Switches = make(map[Coord]Switch)
		}
		l.Switches[n] = Switch{Action: SwitchExit}
		return
	}
}

func (l *Level) SwitchAt(x, y int) (Switch, bool) {
	s, ok := l.Switches[Coord{X: x, Y: y}]
	return s, ok
}

func (l *Level) HasExitSwitch() bool {
	for _, s := range l.Switches {
		if s.Action == SwitchExit {
			return true
		}
	}
	return false
}
