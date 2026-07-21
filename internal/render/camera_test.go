package render

import (
	"math"
	"testing"
)

func TestCameraCentreRayFacesForward(t *testing.T) {
	c := newCamera(0, 1.152) // facing +X
	dx, dy := c.rayDir(100, 200)
	if math.Abs(dx-1) > 1e-9 || math.Abs(dy) > 1e-9 {
		t.Errorf("centre ray = (%.4f,%.4f), want (1,0)", dx, dy)
	}
}

func TestCameraEdgeRaysSpanFOV(t *testing.T) {
	const fov = 1.152
	c := newCamera(0, fov)
	lx, ly := c.rayDir(0, 200)   // left edge
	rx, ry := c.rayDir(200, 200) // right edge (cameraX = +1)
	half := math.Atan2(ry, rx)   // angle of the right edge ray
	if got := 2 * half; math.Abs(got-fov) > 1e-6 {
		t.Errorf("edge-to-edge angle = %.4f, want fov %.4f", got, fov)
	}
	// The two edges are mirror images across the facing direction.
	if math.Abs(math.Atan2(ly, lx)+half) > 1e-9 {
		t.Errorf("edges not symmetric: left=%.4f right=%.4f", math.Atan2(ly, lx), half)
	}
}

func TestCameraRotatesWithAngle(t *testing.T) {
	c := newCamera(math.Pi/2, 1.152) // facing +Y
	dx, dy := c.rayDir(100, 200)
	if math.Abs(dx) > 1e-9 || math.Abs(dy-1) > 1e-9 {
		t.Errorf("centre ray when facing +Y = (%.4f,%.4f), want (0,1)", dx, dy)
	}
}
