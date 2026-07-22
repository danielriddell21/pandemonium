package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

type Vec2 struct {
	X, Y float64
}

const (
	MaxHealth = 100.0

	MaxArmor = 100.0

	armorAbsorb = 1.0 / 3.0

	eyeHeight = 0.5

	fallSpeed = 6.0

	viewRate = 5.0

	baseMaxBullets = 200
	baseMaxShells  = 50
	baseMaxRockets = 50

	overHealMax = 200.0

	invulnDuration  = 20.0
	radSuitDuration = 30.0

	hurtFaceDuration = 0.6
)

type Player struct {
	Pos      Vec2
	Z        float64
	Angle    float64
	Health   float64
	Armor    float64
	Weapon   WeaponKind
	Bullets  int
	Shells   int
	Rockets  int
	Backpack bool
	Keys     map[world.ItemKind]bool

	Berserk    bool
	InvulnTTL  float64
	RadSuitTTL float64

	hurtDir int
	hurtTTL float64
}

func (p Player) FaceDir() int {
	if p.hurtTTL > 0 {
		return p.hurtDir
	}
	return 0
}

func (p Player) Invulnerable() bool { return p.InvulnTTL > 0 }

func (p Player) RadSuited() bool { return p.RadSuitTTL > 0 }

func (p Player) MaxBullets() int { return p.cap(baseMaxBullets) }

func (p Player) MaxShells() int { return p.cap(baseMaxShells) }

func (p Player) MaxRockets() int { return p.cap(baseMaxRockets) }

func (p Player) cap(base int) int {
	if p.Backpack {
		return base * 2
	}
	return base
}

func (p Player) HasKey(k world.ItemKind) bool {
	return p.Keys[k]
}

func (p Player) Dir() Vec2 {
	return Vec2{X: math.Cos(p.Angle), Y: math.Sin(p.Angle)}
}

type Input struct {
	Forward   float64
	Strafe    float64
	Turn      float64
	TurnDelta float64
	Interact  bool
	Attack    bool

	SelectWeapon int
}
