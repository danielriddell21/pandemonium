package world

import "github.com/danielriddell21/crucible/worldgen"

type SwitchAction uint8

const (
	SwitchExit SwitchAction = iota

	SwitchDoor

	SwitchAlarm
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

const alarmPanelSalt = 0x1f83d9ab

func placeAlarmPanel(l *Level) {
	// Its own stream: mounting a panel must not shift what the passes after it
	// roll from the level's generator.
	g := worldgen.NewRNG(l.Seed ^ alarmPanelSalt)
	var walls []Coord
	for y := range l.H {
		for x := range l.W {
			if c := (Coord{X: x, Y: y}); panelWall(l, c) {
				walls = append(walls, c)
			}
		}
	}
	if len(walls) == 0 {
		return
	}
	c := walls[g.IntN(len(walls))]
	l.Set(c.X, c.Y, TileSwitch)
	l.Switches[c] = Switch{Action: SwitchAlarm}
}

func panelWall(l *Level, c Coord) bool {
	if l.At(c.X, c.Y) != TileWall || cheby(c, l.Spawn) < 4 || cheby(c, l.Exit) < 3 {
		return false
	}
	// One open face only, onto plain floor, so the panel is reached head-on
	// and never sits in a doorway.
	faces := 0
	for _, n := range neighbors4(c) {
		switch {
		case !l.InBounds(n.X, n.Y) || l.At(n.X, n.Y) == TileWall:
		case l.At(n.X, n.Y) == TileFloor:
			faces++
		default:
			return false
		}
	}
	return faces == 1
}

func (l *Level) AlarmPanel() (Coord, bool) {
	for c, s := range l.Switches {
		if s.Action == SwitchAlarm {
			return c, true
		}
	}
	return Coord{}, false
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
