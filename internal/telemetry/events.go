// Package telemetry collects structured playtest analytics from a running game:
// per-event records, a rolling per-level path summary, and a cumulative run
// profile across levels. Everything here is plain, JSON-serialisable data plus a
// small fan-out bus; it produces no gameplay effects of its own.
package telemetry

// MarkerInfo describes a tagged structure the player interacted with, including
// (for forks) which branch they took versus the ones they passed over.
type MarkerInfo struct {
	Kind      string   `json:"kind"`
	X         int      `json:"x"`
	Y         int      `json:"y"`
	TakenX    int      `json:"taken_x,omitempty"`
	TakenY    int      `json:"taken_y,omitempty"`
	OptimalX  int      `json:"optimal_x,omitempty"`
	OptimalY  int      `json:"optimal_y,omitempty"`
	Ignored   [][2]int `json:"ignored,omitempty"`
	WrongDoor bool     `json:"wrong_door,omitempty"`
}

// PlayerEvent is a single timestamped record of player activity within a level.
type PlayerEvent struct {
	Tick        uint64      `json:"tick"`
	TimestampMS int64       `json:"timestamp_ms"`
	Type        string      `json:"type"`
	LevelSeed   int64       `json:"level_seed"`
	LevelIndex  int         `json:"level_index"`
	Cell        [2]int      `json:"cell"`
	Marker      *MarkerInfo `json:"marker,omitempty"`
}
