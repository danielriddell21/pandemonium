package render

import (
	"math"
	"testing"
)

// boxGrid is a simple test world: a hollow room with solid walls at x or y of 0
// and size-1, open in between.
type boxGrid struct{ size int }

func (b boxGrid) Solid(x, y int) bool {
	return x <= 0 || y <= 0 || x >= b.size-1 || y >= b.size-1
}

func TestCastRayHitsWallAtKnownDistance(t *testing.T) {
	g := boxGrid{size: 10}
	// From the centre of a 10x10 box at (5,5), looking along +X, the wall face
	// is at x=9, so the perpendicular distance should be ~4.
	h := castRay(g, 5.0, 5.0, 1, 0)
	if h.mapX != 9 {
		t.Errorf("hit mapX = %d, want 9", h.mapX)
	}
	if math.Abs(h.perpDist-4.0) > 1e-6 {
		t.Errorf("perpDist = %v, want 4", h.perpDist)
	}
	if h.side != 0 {
		t.Errorf("side = %d, want 0 (east/west face)", h.side)
	}
}

func TestCastRayNoFisheye(t *testing.T) {
	// Rays fanned across a camera looking at a flat wall must yield perpendicular
	// distances that are equal (a flat wall stays flat), unlike raw ray length.
	g := boxGrid{size: 20}
	cam := newCamera(-math.Pi/2, 1.152) // face -Y toward the top wall
	const width = 64
	var first float64
	for x := 0; x < width; x++ {
		dx, dy := cam.rayDir(x, width)
		h := castRay(g, 10.0, 10.0, dx, dy)
		if x == 0 {
			first = h.perpDist
			continue
		}
		if math.Abs(h.perpDist-first) > 0.5 {
			t.Errorf("column %d perpDist %v deviates from %v (fisheye)", x, h.perpDist, first)
		}
	}
}

func TestCastRayPerpDistNeverZero(t *testing.T) {
	g := boxGrid{size: 4}
	h := castRay(g, 1.01, 1.01, 1, 0)
	if h.perpDist <= 0 {
		t.Errorf("perpDist = %v, must be positive", h.perpDist)
	}
}
