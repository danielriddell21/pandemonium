package telemetry

import (
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

type MarkerInfo struct {
	Kind      string           `json:"kind"`
	KindEnum  world.MarkerKind `json:"-"`
	X         int              `json:"x"`
	Y         int              `json:"y"`
	TakenX    int              `json:"taken_x,omitempty"`
	TakenY    int              `json:"taken_y,omitempty"`
	OptimalX  int              `json:"optimal_x,omitempty"`
	OptimalY  int              `json:"optimal_y,omitempty"`
	Ignored   [][2]int         `json:"ignored,omitempty"`
	WrongDoor bool             `json:"wrong_door,omitempty"`
}

type PlayerEvent struct {
	Tick        uint64              `json:"tick"`
	TimestampMS int64               `json:"timestamp_ms"`
	Type        string              `json:"type"`
	Kind        sim.ObservationKind `json:"-"`
	LevelSeed   int64               `json:"level_seed"`
	LevelIndex  int                 `json:"level_index"`
	Cell        [2]int              `json:"cell"`
	Marker      *MarkerInfo         `json:"marker,omitempty"`
}
