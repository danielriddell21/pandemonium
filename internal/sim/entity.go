package sim

import (
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// EntityKind distinguishes demon behaviours.
type EntityKind uint8

const (
	// Melee demons only harm the player on contact.
	Melee EntityKind = iota
	// Ranged demons hurl projectiles from a distance.
	Ranged
)

// EntityState is a demon's lifecycle phase, used for behaviour and animation.
type EntityState uint8

const (
	// Active demons chase and threaten the player.
	Active EntityState = iota
	// Dying demons are playing their death animation.
	Dying
	// Dead demons are settled corpses.
	Dead
)

// Entity is a billboarded actor in the world — a demon, drawn as a flat sprite
// that always faces the camera.
type Entity struct {
	Pos    Vec2
	Z      float64 // feet height above the base floor, in wall units
	Sprite int     // visual variant
	Kind   EntityKind
	State  EntityState
	Health float64
	Alive  bool // true while Active (targetable, can move and harm)
	Frame  int  // animation frame index for the current state (set by the sim)

	hurt float64 // remaining stagger time after taking a hit, in seconds
	fire float64 // remaining cooldown before a ranged demon shoots again
	anim float64 // animation clock, in seconds
}

// entityCount scales the number of demons with the floor area of the level.
func entityCount(l *world.Level) int {
	floors := 0
	for _, t := range l.Tiles {
		if t.Walkable() {
			floors++
		}
	}
	n := floors / 40
	return min(max(n, 1), 16)
}

// spawnEntities deterministically scatters demons across walkable tiles, keeping
// them clear of the immediate spawn area. Derived from the level seed, so a seed
// always produces the same encounter layout.
func spawnEntities(l *world.Level) []Entity {
	r := rand.New(rand.NewPCG(uint64(l.Seed), 0xA5A5A5A5))

	var floors []world.Coord
	for y := range l.Height {
		for x := range l.Width {
			if l.At(x, y) != world.TileFloor {
				continue
			}
			c := world.Coord{X: x, Y: y}
			if chebyInt(c, l.Spawn) < 3 {
				continue
			}
			floors = append(floors, c)
		}
	}
	if len(floors) == 0 {
		return nil
	}

	// Shuffle deterministically and take the first N distinct tiles.
	r.Shuffle(len(floors), func(i, j int) { floors[i], floors[j] = floors[j], floors[i] })

	n := min(entityCount(l), len(floors))
	ents := make([]Entity, 0, n)
	for i := range n {
		c := floors[i]
		kind := Melee
		if i%3 == 2 { // roughly a third are ranged
			kind = Ranged
		}
		d := newDemon(Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5}, kind)
		d.Z = l.Floor(c.X, c.Y)
		ents = append(ents, d)
	}
	return ents
}

// newDemon builds a fresh demon of the given kind. The visual variant tracks the
// kind so ranged demons read differently from melee ones.
func newDemon(pos Vec2, kind EntityKind) Entity {
	hp := meleeHealth
	sprite := 0
	if kind == Ranged {
		hp = rangedHealth
		sprite = 1
	}
	return Entity{Pos: pos, Sprite: sprite, Kind: kind, State: Active, Health: hp, Alive: true}
}

func chebyInt(a, b world.Coord) int {
	dx, dy := a.X-b.X, a.Y-b.Y
	return max(abs(dx), abs(dy))
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func normalizeAngle(a float64) float64 {
	const twoPi = 2 * math.Pi
	a = math.Mod(a, twoPi)
	if a < 0 {
		a += twoPi
	}
	return a
}
