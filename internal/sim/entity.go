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

	Boss
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

	hurt  float64
	fire  float64
	anim  float64
	alert bool
	wave  bool
}

func (e Entity) Flinching() bool { return e.hurt > 0 }

func (e Entity) Hunting() bool { return e.alert }

const (
	maxDemons   = 24
	bossEscort  = 4
	depthCrowd  = 0.06
	maxDepthMul = 2.0
)

func depthScale(depth int) float64 {
	return min(1+depthCrowd*float64(depth), maxDepthMul)
}

func entityCount(l *world.Level, skill Skill, depth int) int {
	if l.Boss {
		// The overlord's floor opens thin: it calls its own company.
		return max(int(bossEscort*skill.countScale()), 1)
	}
	floors := 0
	for _, t := range l.Tiles {
		if t.Walkable() {
			floors++
		}
	}
	n := int(float64(floors/40) * skill.countScale() * depthScale(depth))
	return min(max(n, 1), maxDemons)
}

var bestiary = [...]EntityKind{Melee, Pinky, Ranged, Ranged, Gunner, Baron, Melee}

func (g *Game) spawnEntities() []Entity {
	l := g.World.Level
	floors := g.demonFloors()

	// Shuffle deterministically and take the first N distinct tiles.
	r := rand.New(rand.NewPCG(uint64(l.Seed), 0xA5A5A5A5))
	r.Shuffle(len(floors), func(i, j int) { floors[i], floors[j] = floors[j], floors[i] })

	n := min(entityCount(l, g.skill, g.depth), len(floors))
	ents := make([]Entity, 0, n+len(l.Barrels)+1)
	for i := range n {
		// A varied bestiary: melee chargers, fast pinkies, fireball imps, hitscan
		// gunners and the occasional armoured baron.
		d := newDemon(centreOf(floors[i]), bestiary[i%len(bestiary)])
		d.Z = l.Floor(floors[i].X, floors[i].Y)
		ents = append(ents, d)
	}
	for _, c := range l.Barrels {
		b := newBarrel(centreOf(c))
		b.Z = l.Floor(c.X, c.Y)
		ents = append(ents, b)
	}
	if l.Boss {
		ents = append(ents, g.newBoss())
	}
	return ents
}

func (g *Game) demonFloors() []world.Coord {
	l := g.World.Level
	var floors []world.Coord
	for y := range l.H {
		for x := range l.W {
			if l.At(x, y) != world.TileFloor {
				continue
			}
			c := world.Coord{X: x, Y: y}
			if chebyInt(c, l.Spawn) < 3 || (l.Boss && c == l.BossAt) || g.overlooksSpawn(c) {
				continue
			}
			floors = append(floors, c)
		}
	}
	return floors
}

const spawnCalm = detectRadius + 1.5

func (g *Game) overlooksSpawn(c world.Coord) bool {
	// Nothing may open fire the moment a level begins: a demon starts either
	// out of sight of the spawn or too far off to have noticed it.
	spawn := centreOf(g.World.Level.Spawn)
	return dist(centreOf(c), spawn) < spawnCalm && losClear(g.World, centreOf(c), spawn)
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
	case Boss:
		sprite = 5
	}
	return Entity{Pos: pos, Sprite: sprite, Kind: kind, State: Active, Health: hp, Alive: true}
}

func cellOf(p Vec2) world.Coord {
	return world.Coord{X: int(math.Floor(p.X)), Y: int(math.Floor(p.Y))}
}

func centreOf(c world.Coord) Vec2 {
	return Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5}
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
