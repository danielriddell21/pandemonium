package sim

import (
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/pandemonium/internal/world"
)

type EntityKind uint8

const (
	Melee EntityKind = iota

	Ranged

	Gunner

	Barrel

	Pinky

	Baron
)

type EntityState uint8

const (
	Active EntityState = iota

	Dying

	Dead
)

type Entity struct {
	Pos    Vec2
	Z      float64
	Sprite int
	Kind   EntityKind
	State  EntityState
	Health float64
	Alive  bool
	Frame  int

	hurt float64
	fire float64
	anim float64
}

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

func (g *Game) spawnEntities() []Entity {
	l := g.World.Level
	r := rand.New(rand.NewPCG(uint64(l.Seed), 0xA5A5A5A5))

	var floors []world.Coord
	for y := range l.H {
		for x := range l.W {
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

func newBarrel(pos Vec2) Entity {
	return Entity{Pos: pos, Sprite: 0, Kind: Barrel, State: Active, Health: barrelHealth, Alive: true}
}

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

func ranges(kind EntityKind) bool { return kind == Ranged || kind == Baron }

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
