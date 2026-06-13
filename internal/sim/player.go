// Package sim is the headless game simulation: it advances player movement,
// collision, entities and interaction one fixed step at a time. It depends only
// on the world package and knows nothing about how the game is drawn.
package sim

import "math"

// Vec2 is a 2D vector in world space, measured in tiles.
type Vec2 struct {
	X, Y float64
}

// MaxHealth is the player's starting and maximum health.
const MaxHealth = 100.0

// Player holds the camera-bearing actor's position, facing and health.
type Player struct {
	Pos    Vec2
	Angle  float64 // radians; 0 points along +X
	Health float64
}

// Dir returns the unit vector the player is facing.
func (p Player) Dir() Vec2 {
	return Vec2{X: math.Cos(p.Angle), Y: math.Sin(p.Angle)}
}

// Input is the per-tick set of movement intents, each normalised to roughly
// [-1, 1]. It is produced by the front-end and consumed by Tick.
type Input struct {
	Forward   float64 // +forward / -backward
	Strafe    float64 // +right / -left
	Turn      float64 // rate-based turn: +clockwise / -counter-clockwise
	TurnDelta float64 // direct turn applied this tick, in radians (mouse-look)
	Interact  bool    // act on an adjacent door this tick
	Attack    bool    // strike straight ahead this tick
}
