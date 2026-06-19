// Package world generates and represents game levels as 2D tile grids. It is
// pure data and logic: it has no knowledge of rendering or simulation, imports
// no graphics libraries, and is fully testable headlessly. Levels are produced
// deterministically from a seed.
package world

// TileType enumerates the kinds of cell a level grid can contain.
type TileType uint8

const (
	// TileFloor is open, walkable space.
	TileFloor TileType = iota
	// TileWall is solid and blocks movement and sight.
	TileWall
	// TileDoor is a passage that blocks until opened.
	TileDoor
	// TileSpawn marks where the player starts. It is walkable.
	TileSpawn
	// TileExit marks the floor in front of the exit switch. It is walkable.
	TileExit
	// TileSwitch is a wall-mounted switch the player presses with use. It is solid
	// like a wall; what it does is recorded in Level.Switches.
	TileSwitch
)

// Walkable reports whether an actor can stand on this tile type. Doors are
// considered walkable; whether a specific door is currently passable is tracked
// separately by the simulation.
func (t TileType) Walkable() bool {
	switch t {
	case TileFloor, TileSpawn, TileExit, TileDoor:
		return true
	default:
		return false
	}
}

// Rune returns a compact character for debug/ASCII rendering of a grid.
func (t TileType) Rune() rune {
	switch t {
	case TileWall:
		return '#'
	case TileDoor:
		return '+'
	case TileSpawn:
		return 'S'
	case TileExit:
		return 'E'
	case TileSwitch:
		return '/'
	default:
		return '.'
	}
}
