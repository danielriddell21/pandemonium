package render

import "math"

// camera is the pair of vectors a raycaster needs: the unit direction the player
// faces and the camera plane perpendicular to it, whose half-length encodes the
// field of view.
type camera struct {
	dirX, dirY     float64
	planeX, planeY float64
}

// newCamera builds a camera from a facing angle and field of view. The plane
// length is tan(fov/2), so a ray through the screen edge subtends fov/2.
func newCamera(angle, fov float64) camera {
	dirX, dirY := math.Cos(angle), math.Sin(angle)
	planeLen := math.Tan(fov / 2)
	return camera{
		dirX:   dirX,
		dirY:   dirY,
		planeX: -dirY * planeLen,
		planeY: dirX * planeLen,
	}
}

// rayDir returns the world-space direction of the ray through screen column x of
// width columns. cameraX runs from -1 at the left edge to +1 at the right.
func (c camera) rayDir(x, width int) (float64, float64) {
	cameraX := 2*float64(x)/float64(width) - 1
	return c.dirX + c.planeX*cameraX, c.dirY + c.planeY*cameraX
}
