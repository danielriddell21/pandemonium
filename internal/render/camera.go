package render

import "math"

type camera struct {
	dirX, dirY     float64
	planeX, planeY float64
}

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

func (c camera) rayDir(x, width int) (float64, float64) {
	cameraX := 2*float64(x)/float64(width) - 1
	return c.dirX + c.planeX*cameraX, c.dirY + c.planeY*cameraX
}
