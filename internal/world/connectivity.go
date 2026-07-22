package world

import (
	"github.com/danielriddell21/crucible/level"
	"github.com/danielriddell21/crucible/worldgen"
)

type solidFn = func(Coord) bool

func blocksClosed(l *Level) solidFn {
	return func(c Coord) bool { return l.Solid(c.X, c.Y) }
}

func blocksWalls(l *Level) solidFn {
	return func(c Coord) bool { return !l.At(c.X, c.Y).Walkable() }
}

// climbable gates each flood step by the engine's height rules, so
// reachability respects the ledges and lifts heights introduce.
func climbable(l *Level) func(from, to Coord) bool {
	return func(from, to Coord) bool {
		return l.StepOK(from, to, level.DefaultMaxStep, level.DefaultMinHeadroom)
	}
}

// Reachable reports whether dst can be reached from src with closed doors
// blocking, under the height rules.
func Reachable(l *Level, src, dst Coord) bool {
	return reachable(l, src, dst, blocksClosed(l))
}

func reachable(l *Level, src, dst Coord, solid solidFn) bool {
	return worldgen.FloodDist(l.W, l.H, src, solid, climbable(l)).At(dst) >= 0
}

// StepsBetween returns the height-gated step distance from src to dst, or -1
// when dst is unreachable.
func StepsBetween(l *Level, src, dst Coord) int {
	return worldgen.FloodDist(l.W, l.H, src, blocksWalls(l), climbable(l)).At(dst)
}

func neighbors4(c Coord) [4]Coord {
	return worldgen.Neighbors4(c)
}
