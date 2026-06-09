package world

// Reachable reports whether dst can be reached from src by moving through
// non-solid tiles in the four cardinal directions. It is a breadth-first flood
// fill and is the guarantee that backs Generate: every level it returns has a
// path from spawn to exit.
func Reachable(l *Level, src, dst Coord) bool {
	if !l.InBounds(src.X, src.Y) || !l.InBounds(dst.X, dst.Y) {
		return false
	}
	if l.Solid(src.X, src.Y) || l.Solid(dst.X, dst.Y) {
		return false
	}

	visited := make([]bool, l.Width*l.Height)
	queue := []Coord{src}
	visited[src.Y*l.Width+src.X] = true

	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == dst {
			return true
		}
		for _, n := range neighbors4(c) {
			if !l.InBounds(n.X, n.Y) || l.Solid(n.X, n.Y) {
				continue
			}
			idx := n.Y*l.Width + n.X
			if visited[idx] {
				continue
			}
			visited[idx] = true
			queue = append(queue, n)
		}
	}
	return false
}

// neighbors4 returns the four cardinal neighbours of c.
func neighbors4(c Coord) [4]Coord {
	return [4]Coord{
		{c.X + 1, c.Y},
		{c.X - 1, c.Y},
		{c.X, c.Y + 1},
		{c.X, c.Y - 1},
	}
}
