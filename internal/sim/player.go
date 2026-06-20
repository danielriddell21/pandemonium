// Package sim is the headless game simulation: it advances player movement,
// collision, entities and interaction one fixed step at a time. It depends only
// on the world package and knows nothing about how the game is drawn.
package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// Vec2 is a 2D vector in world space, measured in tiles.
type Vec2 struct {
	X, Y float64
}

const (
	// MaxHealth is the player's starting and maximum health.
	MaxHealth = 100.0
	// MaxArmor is the most armour the player can carry.
	MaxArmor = 100.0
	// armorAbsorb is the fraction of incoming damage soaked by armour while the
	// player has any, matching DOOM's green-armour behaviour.
	armorAbsorb = 1.0 / 3.0

	// eyeHeight is how far the camera sits above the feet, in wall units. At the
	// base floor this puts the horizon at mid-screen, as before heights existed.
	eyeHeight = 0.5
	// fallSpeed is how fast a body drops toward its floor, in wall units/second.
	fallSpeed = 6.0
	// viewRate is how quickly the camera height eases toward the body's height,
	// so stairs read as steps rather than jolts.
	viewRate = 5.0

	// Base ammo capacities; a backpack doubles them (see Player.MaxBullets etc.).
	baseMaxBullets = 200
	baseMaxShells  = 50
	baseMaxRockets = 50
)

// Player holds the camera-bearing actor's position, facing, height, health,
// armour and arsenal, plus the keycards it has collected.
type Player struct {
	Pos      Vec2
	Z        float64 // feet height above the base floor, in wall units
	Angle    float64 // radians; 0 points along +X
	Health   float64
	Armor    float64
	Weapon   WeaponKind
	Bullets  int
	Shells   int
	Rockets  int
	Backpack bool // doubles ammo capacity once collected
	Keys     map[world.ItemKind]bool
}

// MaxBullets is the player's current bullet capacity (doubled by a backpack).
func (p Player) MaxBullets() int { return p.cap(baseMaxBullets) }

// MaxShells is the player's current shell capacity (doubled by a backpack).
func (p Player) MaxShells() int { return p.cap(baseMaxShells) }

// MaxRockets is the player's current rocket capacity (doubled by a backpack).
func (p Player) MaxRockets() int { return p.cap(baseMaxRockets) }

func (p Player) cap(base int) int {
	if p.Backpack {
		return base * 2
	}
	return base
}

// HasKey reports whether the player holds the given keycard.
func (p Player) HasKey(k world.ItemKind) bool {
	return p.Keys[k]
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
	Attack    bool    // fire the current weapon this tick
	// SelectWeapon switches weapon when non-zero: 1=fists, 2=pistol, 3=shotgun,
	// 4=chaingun, 5=rocket launcher.
	SelectWeapon int
}
