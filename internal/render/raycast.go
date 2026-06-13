package render

import "math"

// grid is the minimal view of the world a raycaster needs: whether a cell blocks
// sight. sim.World satisfies it, keeping the raycaster decoupled and testable.
type grid interface {
	Solid(x, y int) bool
}

// hit is the result of casting one ray: the wall cell struck, the perpendicular
// distance to it (corrected to avoid fisheye), which face was hit, and where
// along the wall it landed (wallX in [0,1), for texture mapping).
type hit struct {
	perpDist   float64
	mapX, mapY int
	side       int // 0 = vertical (east/west) face, 1 = horizontal (north/south)
	wallX      float64
}

// castRay walks a ray from (posX, posY) in direction (dirX, dirY) through the
// grid using a DDA, stepping exactly cell-to-cell until it hits a solid tile. It
// returns the perpendicular distance to that wall, which keeps straight walls
// flat on screen.
func castRay(g grid, posX, posY, dirX, dirY float64) hit {
	mapX, mapY := int(math.Floor(posX)), int(math.Floor(posY))

	// Distance the ray travels to cross one full cell in each axis.
	deltaX := math.Inf(1)
	if dirX != 0 {
		deltaX = math.Abs(1 / dirX)
	}
	deltaY := math.Inf(1)
	if dirY != 0 {
		deltaY = math.Abs(1 / dirY)
	}

	// Step direction and the distance from the start to the first grid line.
	var stepX, stepY int
	var sideDistX, sideDistY float64
	if dirX < 0 {
		stepX = -1
		sideDistX = (posX - float64(mapX)) * deltaX
	} else {
		stepX = 1
		sideDistX = (float64(mapX) + 1 - posX) * deltaX
	}
	if dirY < 0 {
		stepY = -1
		sideDistY = (posY - float64(mapY)) * deltaY
	} else {
		stepY = 1
		sideDistY = (float64(mapY) + 1 - posY) * deltaY
	}

	// Advance to the nearer grid line each iteration until a wall is hit.
	side := 0
	for range maxDDASteps {
		if sideDistX < sideDistY {
			sideDistX += deltaX
			mapX += stepX
			side = 0
		} else {
			sideDistY += deltaY
			mapY += stepY
			side = 1
		}
		if g.Solid(mapX, mapY) {
			break
		}
	}

	// Perpendicular distance: subtract the last increment so we measure to the
	// wall face rather than past it. This is the projection onto the camera
	// direction, which removes the fisheye bow.
	perp := sideDistX - deltaX
	if side == 1 {
		perp = sideDistY - deltaY
	}
	if perp < 1e-6 {
		perp = 1e-6
	}

	// Where along the struck wall the ray landed, used to pick a texture column.
	var wallX float64
	if side == 0 {
		wallX = posY + perp*dirY
	} else {
		wallX = posX + perp*dirX
	}
	wallX -= math.Floor(wallX)

	return hit{perpDist: perp, mapX: mapX, mapY: mapY, side: side, wallX: wallX}
}

// maxDDASteps bounds DDA iterations so a ray that somehow escapes the bounded
// world can never loop forever.
const maxDDASteps = 4096
