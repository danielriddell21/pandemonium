package world

import "github.com/danielriddell21/crucible/level"

// TileType is the engine's spatial tile vocabulary from crucible/level.
type TileType = level.Tile

const (
	TileFloor  = level.TileFloor
	TileWall   = level.TileWall
	TileDoor   = level.TileDoor
	TileSpawn  = level.TileSpawn
	TileExit   = level.TileExit
	TileSwitch = level.TileSwitch
)
