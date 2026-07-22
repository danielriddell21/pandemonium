package world

type TileType uint8

const (
	TileFloor TileType = iota

	TileWall

	TileDoor

	TileSpawn

	TileExit

	TileSwitch
)

func (t TileType) Walkable() bool {
	switch t {
	case TileFloor, TileSpawn, TileExit, TileDoor:
		return true
	default:
		return false
	}
}

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
