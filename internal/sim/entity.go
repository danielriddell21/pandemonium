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
	// Gunner demons fire instantly (hitscan) on a cooldown, the way DOOM's
	// zombiemen and shotgun guys do — no projectile to dodge.
	Gunner
	// Barrel is a static explosive: it doesn't move or attack, but bursts when
	// destroyed, splashing damage onto whatever is near (chaining other barrels).
	Barrel
	// Pinky is a fast, low-health melee charger.
	Pinky
	// Baron is a slow, heavily-armoured demon that hurls big, slow fireballs.
	Baron
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

// entityCount scales the number of demons with the floor area of the level and
// the skill: harder skills pack more demons in (and allow a higher ceiling).
func entityCount(l *world.Level, skill Skill) int {
	floors := 0
	for _, t := range l.Tiles {
		if t.Walkable() {
			floors++
		}
	}
	n := int(float64(floors/40) * skill.countScale())
	return min(max(n, 1), 24)
}

// spawnEntities deterministically scatters demons across walkable tiles, keeping
// them clear of the immediate spawn area. Derived from the level seed, so a seed
// always produces the same encounter layout (the count scaled by the skill).
func (g *Game) spawnEntities() []Entity {
	l := g.World.Level
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

	n := min(entityCount(l, g.skill), len(floors))
	ents := make([]Entity, 0, n)
	for i := range n {
		c := floors[i]
		// A varied bestiary: melee chargers, fast pinkies, fireball imps, hitscan
		// gunners and the occasional armoured baron.
		kind := Melee
		switch i % 7 {
		case 1:
			kind = Pinky
		case 2, 3:
			kind = Ranged
		case 4:
			kind = Gunner
		case 5:
			kind = Baron
		}
		d := newDemon(Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5}, kind)
		d.Z = l.Floor(c.X, c.Y)
		ents = append(ents, d)
	}
	for _, c := range l.Barrels {
		b := newBarrel(Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5})
		b.Z = l.Floor(c.X, c.Y)
		ents = append(ents, b)
	}
	return ents
}

// newBarrel builds a static explosive barrel.
func newBarrel(pos Vec2) Entity {
	return Entity{Pos: pos, Sprite: 0, Kind: Barrel, State: Active, Health: barrelHealth, Alive: true}
}

// newDemon builds a fresh demon of the given kind. The visual variant tracks the
// kind so each behaviour reads as a distinct silhouette.
func newDemon(pos Vec2, kind EntityKind) Entity {
	hp, sprite := meleeHealth, 0
	switch kind {
	case Ranged:
		hp, sprite = rangedHealth, 1
	case Gunner:
		hp, sprite = gunnerHealth, 2
	case Pinky:
		hp, sprite = pinkyHealth, 3
	case Baron:
		hp, sprite = baronHealth, 4
	}
	return Entity{Pos: pos, Sprite: sprite, Kind: kind, State: Active, Health: hp, Alive: true}
}

// demonSpeed returns a demon's chase speed; pinkies rush, barons lumber.
func demonSpeedFor(kind EntityKind) float64 {
	switch kind {
	case Pinky:
		return pinkySpeed
	case Baron:
		return baronSpeed
	default:
		return demonSpeed
	}
}

// ranges reports whether a kind attacks with projectiles.
func ranges(kind EntityKind) bool { return kind == Ranged || kind == Baron }

// countDemons counts the entities that are actual demons (excluding barrels), so
// the kill tally's denominator isn't inflated by explosive props.
func countDemons(ents []Entity) int {
	n := 0
	for _, e := range ents {
		if e.Kind != Barrel {
			n++
		}
	}
	return n
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
