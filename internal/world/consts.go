package world

import "github.com/danielriddell21/crucible/level"

const (
	// MaxStep is the largest floor rise an agent can climb in one step,
	// from crucible/level.
	MaxStep = level.DefaultMaxStep
	// MinHeadroom is the smallest floor-to-ceiling gap an agent needs,
	// from crucible/level.
	MinHeadroom = level.DefaultMinHeadroom
	// NumThemes is how many wall themes the generator paints rooms with.
	NumThemes = 3
)
