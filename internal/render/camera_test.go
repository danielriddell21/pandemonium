package render

import (
	"math"
	"testing"

	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/raycast"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

// testCam builds a raycast camera at the game's player position, the way the
// renderer does, for the drawing tests.
func testCam(g *sim.Game, angle, fov float64) raycast.Camera {
	return raycast.NewCamera(geom.Vec2{X: g.Player.Pos.X, Y: g.Player.Pos.Y}, angle, fov)
}

// These check pandemonium's field of view through crucible/raycast's camera;
// the camera math itself is covered by crucible's own tests.

func TestCameraCentreRayFacesForward(t *testing.T) {
	c := raycast.NewCamera(geom.Vec2{}, 0, 1.152) // facing +X
	dx, dy := c.RayDir(100, 200)
	if math.Abs(dx-1) > 1e-9 || math.Abs(dy) > 1e-9 {
		t.Errorf("centre ray = (%.4f,%.4f), want (1,0)", dx, dy)
	}
}

func TestCameraEdgeRaysSpanFOV(t *testing.T) {
	const fov = 1.152
	c := raycast.NewCamera(geom.Vec2{}, 0, fov)
	lx, ly := c.RayDir(0, 200)   // left edge
	rx, ry := c.RayDir(200, 200) // right edge (cameraX = +1)
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
	c := raycast.NewCamera(geom.Vec2{}, math.Pi/2, 1.152) // facing +Y
	dx, dy := c.RayDir(100, 200)
	if math.Abs(dx) > 1e-9 || math.Abs(dy-1) > 1e-9 {
		t.Errorf("centre ray when facing +Y = (%.4f,%.4f), want (0,1)", dx, dy)
	}
}
