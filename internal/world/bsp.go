package world

// This file implements binary space partitioning of the grid into rooms joined
// by corridors. The map is recursively cut into sub-regions; each leaf region
// gets a room carved into it; sibling regions are then linked with L-shaped
// corridors so the whole tree is connected.

const (
	// minLeaf is the smallest dimension a region may have after a split. A
	// region must be at least 2*minLeaf along an axis to be split on that axis.
	minLeaf = 7
	// maxLeaf is the size above which a region is always split; below it,
	// splitting becomes probabilistic to vary room sizes.
	maxLeaf = 16
	// minRoom is the smallest room dimension carved into a leaf.
	minRoom = 3
	// roomPad keeps carved rooms one tile inside their region so adjacent rooms
	// never merge into each other.
	roomPad = 1
	// maxDepth caps recursion; size guards usually stop splitting first.
	maxDepth = 7
)

// rect is an axis-aligned rectangle in grid space.
type rect struct {
	x, y, w, h int
}

// center returns the rectangle's centre cell.
func (r rect) center() Coord {
	return Coord{X: r.x + r.w/2, Y: r.y + r.h/2}
}

// bspNode is one region in the partition tree. Internal nodes have two children;
// leaf nodes carry a carved room.
type bspNode struct {
	bounds      rect
	left, right *bspNode
	room        *rect
}

// leaf reports whether this node has no children.
func (n *bspNode) leaf() bool {
	return n.left == nil && n.right == nil
}

// split recursively partitions a region into children.
func (g *rng) split(n *bspNode, depth int) {
	if depth >= maxDepth {
		return
	}
	w, h := n.bounds.w, n.bounds.h
	canV := w >= 2*minLeaf // vertical cut -> left/right
	canH := h >= 2*minLeaf // horizontal cut -> top/bottom
	if !canV && !canH {
		return
	}
	// Small enough regions stop splitting most of the time, for size variety.
	if w <= maxLeaf && h <= maxLeaf && g.chance(0.3) {
		return
	}

	vertical := chooseSplitAxis(g, w, h, canV, canH)
	if vertical {
		at := g.between(minLeaf, w-minLeaf)
		n.left = &bspNode{bounds: rect{n.bounds.x, n.bounds.y, at, h}}
		n.right = &bspNode{bounds: rect{n.bounds.x + at, n.bounds.y, w - at, h}}
	} else {
		at := g.between(minLeaf, h-minLeaf)
		n.left = &bspNode{bounds: rect{n.bounds.x, n.bounds.y, w, at}}
		n.right = &bspNode{bounds: rect{n.bounds.x, n.bounds.y + at, w, h - at}}
	}
	g.split(n.left, depth+1)
	g.split(n.right, depth+1)
}

// chooseSplitAxis picks a cut orientation, preferring to split the longer side
// so rooms stay reasonably square.
func chooseSplitAxis(g *rng, w, h int, canV, canH bool) bool {
	switch {
	case canV && !canH:
		return true
	case canH && !canV:
		return false
	case w > h:
		return true
	case h > w:
		return false
	default:
		return g.chance(0.5)
	}
}

// carveRooms walks to every leaf and carves a randomly sized/placed room inside
// its region, recording it on the node and stamping floor tiles into the level.
func (g *rng) carveRooms(n *bspNode, l *Level) {
	if !n.leaf() {
		if n.left != nil {
			g.carveRooms(n.left, l)
		}
		if n.right != nil {
			g.carveRooms(n.right, l)
		}
		return
	}
	b := n.bounds
	maxW, maxH := b.w-2*roomPad, b.h-2*roomPad
	if maxW < minRoom || maxH < minRoom {
		return
	}
	rw := g.between(minRoom, maxW)
	rh := g.between(minRoom, maxH)
	rx := b.x + roomPad + g.intn(maxW-rw+1)
	ry := b.y + roomPad + g.intn(maxH-rh+1)
	room := rect{rx, ry, rw, rh}
	n.room = &room
	for y := ry; y < ry+rh; y++ {
		for x := rx; x < rx+rw; x++ {
			l.set(x, y, TileFloor)
		}
	}
}

// connect links the rooms of a node's two subtrees with a corridor, bottom-up,
// returning a representative room for the subtree so parents can keep linking.
func (g *rng) connect(n *bspNode, l *Level) *rect {
	if n.room != nil {
		return n.room
	}
	var lr, rr *rect
	if n.left != nil {
		lr = g.connect(n.left, l)
	}
	if n.right != nil {
		rr = g.connect(n.right, l)
	}
	if lr != nil && rr != nil {
		g.carveCorridor(lr.center(), rr.center(), l)
	}
	if lr != nil {
		return lr
	}
	return rr
}

// carveCorridor digs an L-shaped corridor between two cells, choosing which leg
// to dig first at random.
func (g *rng) carveCorridor(a, b Coord, l *Level) {
	if g.chance(0.5) {
		carveH(l, a.X, b.X, a.Y)
		carveV(l, a.Y, b.Y, b.X)
	} else {
		carveV(l, a.Y, b.Y, a.X)
		carveH(l, a.X, b.X, b.Y)
	}
}

func carveH(l *Level, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if l.At(x, y) == TileWall {
			l.set(x, y, TileFloor)
		}
	}
}

func carveV(l *Level, y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if l.At(x, y) == TileWall {
			l.set(x, y, TileFloor)
		}
	}
}

// collectRooms returns every carved room in deterministic in-order, so spawn and
// exit selection is reproducible.
func collectRooms(n *bspNode) []rect {
	if n == nil {
		return nil
	}
	if n.room != nil {
		return []rect{*n.room}
	}
	return append(collectRooms(n.left), collectRooms(n.right)...)
}
