package sim

import (
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// Entity is a billboarded actor in the world — a demon, drawn as a flat sprite
// that always faces the camera.
type Entity struct {
	Pos    Vec2
	Sprite int
	Alive  bool
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
		ents = append(ents, Entity{
			Pos:    Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5},
			Sprite: r.IntN(spriteVariants),
			Alive:  true,
		})
	}
	return ents
}

// spriteVariants is how many demon sprite variants exist; placement picks among
// them so a crowd isn't visually identical.
const spriteVariants = 2

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
