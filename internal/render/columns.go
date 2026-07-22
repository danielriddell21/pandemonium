package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

const maxDDASteps = 4096

const ceilingDim = 0.85

func drawScene(fb []byte, zbuf, loZ, loH []float64, loRow []int, g *sim.Game, cam camera, cfg Config, tx *textureSet, gloom float64) {
	eyeZ := g.EyeZ()
	for x := range cfg.Width {
		zbuf[x], loZ[x], loH[x], loRow[x] = drawColumn(fb, g, cam, cfg, tx, x, eyeZ, gloom)
	}
}

func drawColumn(fb []byte, g *sim.Game, cam camera, cfg Config, tx *textureSet, x int, eyeZ, gloom float64) (closeDist, loZ, loH float64, loRow int) {
	loZ = math.Inf(1)
	w, h := cfg.Width, cfg.Height
	fh := float64(h)
	px, py := g.Player.Pos.X, g.Player.Pos.Y
	dx, dy := cam.rayDir(x, w)

	// row projects a world height z at perpendicular distance d to a screen row.
	row := func(z, d float64) int {
		return int(fh/2 + (eyeZ-z)*fh/d)
	}

	// DDA state, walking every tile boundary along the ray.
	mapX, mapY := int(math.Floor(px)), int(math.Floor(py))
	deltaX, deltaY := math.Inf(1), math.Inf(1)
	if dx != 0 {
		deltaX = math.Abs(1 / dx)
	}
	if dy != 0 {
		deltaY = math.Abs(1 / dy)
	}
	var stepX, stepY int
	var sideDistX, sideDistY float64
	if dx < 0 {
		stepX, sideDistX = -1, (px-float64(mapX))*deltaX
	} else {
		stepX, sideDistX = 1, (float64(mapX)+1-px)*deltaX
	}
	if dy < 0 {
		stepY, sideDistY = -1, (py-float64(mapY))*deltaY
	} else {
		stepY, sideDistY = 1, (float64(mapY)+1-py)*deltaY
	}

	yTop, yBot := 0, h-1
	aX, aY := mapX, mapY // the tile currently being left (for its floor texture)
	aFloor := g.World.FloorAt(mapX, mapY)
	aCeil := g.World.CeilAt(mapX, mapY)

	for range maxDDASteps {
		// Advance to the next boundary.
		var d float64
		var side int
		if sideDistX < sideDistY {
			d, side = sideDistX, 0
			sideDistX += deltaX
			mapX += stepX
		} else {
			d, side = sideDistY, 1
			sideDistY += deltaY
			mapY += stepY
		}
		if d < 1e-6 {
			d = 1e-6
		}

		floorEdge, ceilEdge := fillDepartedSurfaces(fb, cfg, g, tx, x, aX, aY, yTop, yBot, aFloor, aCeil, d, eyeZ, gloom, px, py, dx, dy)

		// The texture column for any face on this boundary, themed by the room
		// it's seen from and lit by the cell it faces.
		texX, tex := boundaryTexture(g, tx, mapX, mapY, side, d, px, py, dx, dy, g.World.Level.ThemeAt(aX, aY))
		bLight := g.World.Level.LightAt(mapX, mapY) * gloom

		bFloor, bCeil, opaque := blockFaces(g, mapX, mapY, g.World.FloorAt(mapX, mapY), g.World.CeilAt(mapX, mapY))
		if opaque {
			drawWallSpan(fb, cfg, x, max(yTop, ceilEdge+1), min(yBot, floorEdge), d, eyeZ, texX, tex, side, bLight)
			return d, loZ, loH, loRow
		}

		loZ, loH, loRow = drawStepFaces(fb, cfg, x, yTop, yBot, floorEdge, ceilEdge, d, eyeZ, texX, tex, side, bLight, aFloor, aCeil, bFloor, bCeil, loZ, loH, loRow)

		yBot = min(yBot, row(max(aFloor, bFloor), d))
		yTop = max(yTop, row(min(aCeil, bCeil), d)+1)
		if yTop > yBot {
			return d, loZ, loH, loRow
		}
		aFloor, aCeil = bFloor, bCeil
		aX, aY = mapX, mapY
	}
	return math.MaxFloat64, loZ, loH, loRow
}

func fillDepartedSurfaces(fb []byte, cfg Config, g *sim.Game, tx *textureSet, x, aX, aY, yTop, yBot int, aFloor, aCeil, d, eyeZ, gloom, px, py, dx, dy float64) (floorEdge, ceilEdge int) {
	fh := float64(cfg.Height)
	row := func(z float64) int { return int(fh/2 + (eyeZ-z)*fh/d) }
	aLight := g.World.Level.LightAt(aX, aY) * gloom
	floorEdge = row(aFloor)
	ftex := tx.floor
	if g.World.HazardAt(aX, aY) > 0 {
		ftex = tx.nukage
		if g.World.Level.HazardKindAt(aX, aY) == world.HazardLava {
			ftex = tx.lava
		}
	}
	fillFloorSpan(fb, cfg, x, max(yTop, floorEdge+1), yBot, aFloor, eyeZ, px, py, dx, dy, ftex, aLight)
	ceilEdge = row(aCeil)
	if g.World.Level.SkyAt(aX, aY) {
		// Open air overhead: a bright, distance-independent sky rather than
		// distance-shaded stone (still dimming with the run's gloom).
		fillSkySpan(fb, cfg, x, yTop, min(yBot, ceilEdge), gloom)
	} else {
		fillCeilSpan(fb, cfg, x, yTop, min(yBot, ceilEdge), aCeil, eyeZ, px, py, dx, dy, tx.ceiling, aLight)
	}
	return floorEdge, ceilEdge
}

func blockFaces(g *sim.Game, mapX, mapY int, bFloor, bCeil float64) (nf, nc float64, opaque bool) {
	if !g.World.Solid(mapX, mapY) {
		return bFloor, bCeil, false
	}
	if top := g.World.Level.WallTopAt(mapX, mapY); top > 0 {
		// A low wall: a solid block we can see over. Treat it like an unclimbable
		// step up to its top, then keep walking the ray so the room beyond is
		// drawn above it.
		return top, 1, false
	}
	return bFloor, bCeil, true
}

func drawStepFaces(fb []byte, cfg Config, x, yTop, yBot, floorEdge, ceilEdge int, d, eyeZ float64, texX int, tex *texture, side int, bLight, aFloor, aCeil, bFloor, bCeil, loZ, loH float64, loRow int) (float64, float64, int) {
	fh := float64(cfg.Height)
	row := func(z float64) int { return int(fh/2 + (eyeZ-z)*fh/d) }
	if bFloor > aFloor { // rising step face — a low wall, stair, lift or ledge
		drawWallSpan(fb, cfg, x, max(yTop, row(bFloor)+1), min(yBot, floorEdge), d, eyeZ, texX, tex, side, bLight)
		if bFloor > loH {
			loH, loZ, loRow = bFloor, d, row(bFloor)
		}
	}
	if bCeil < aCeil { // dropping ceiling face
		drawWallSpan(fb, cfg, x, max(yTop, ceilEdge+1), min(yBot, row(bCeil)), d, eyeZ, texX, tex, side, bLight)
	}
	return loZ, loH, loRow
}

func boundaryTexture(g *sim.Game, tx *textureSet, mapX, mapY, side int, d, px, py, dx, dy float64, theme uint8) (int, *texture) {
	var wallX float64
	if side == 0 {
		wallX = py + d*dy
	} else {
		wallX = px + d*dx
	}
	wallX -= math.Floor(wallX)

	tex := tx.walls[int(theme)%len(tx.walls)] // themed by the room being viewed from
	switch g.World.Level.At(mapX, mapY) {
	case world.TileDoor:
		tex = tx.door
	case world.TileSwitch:
		tex = tx.switchTex
	}
	texX := int(wallX * float64(tex.w))
	if texX >= tex.w {
		texX = tex.w - 1
	}
	if (side == 0 && dx > 0) || (side == 1 && dy < 0) {
		texX = tex.w - 1 - texX
	}
	return texX, tex
}

func drawWallSpan(fb []byte, cfg Config, x, y0, y1 int, d, eyeZ float64, texX int, tex *texture, side int, light float64) {
	fh := float64(cfg.Height)
	for y := y0; y <= y1; y++ {
		z := eyeZ - (float64(y)-fh/2)*d/fh
		v := int((1 - z) * float64(tex.h)) // tex.at wraps, tiling tall faces
		setPixel(fb, cfg.Width, x, y, scaleColor(shade(tex.at(texX, v), d, side), light))
	}
}

func fillFloorSpan(fb []byte, cfg Config, x, y0, y1 int, z, eyeZ, px, py, dx, dy float64, tex *texture, light float64) {
	fh := float64(cfg.Height)
	for y := y0; y <= y1; y++ {
		p := float64(y) - fh/2
		if p < 0.5 {
			p = 0.5
		}
		rowDist := (eyeZ - z) * fh / p
		sampleFlat(fb, cfg, x, y, rowDist, px, py, dx, dy, tex, shadeFactor(rowDist)*light)
	}
}

func fillCeilSpan(fb []byte, cfg Config, x, y0, y1 int, z, eyeZ, px, py, dx, dy float64, tex *texture, light float64) {
	fh := float64(cfg.Height)
	for y := y0; y <= y1; y++ {
		p := fh/2 - float64(y)
		if p < 0.5 {
			p = 0.5
		}
		rowDist := (z - eyeZ) * fh / p
		sampleFlat(fb, cfg, x, y, rowDist, px, py, dx, dy, tex, shadeFactor(rowDist)*ceilingDim*light)
	}
}

func fillSkySpan(fb []byte, cfg Config, x, y0, y1 int, gloom float64) {
	half := float64(cfg.Height) / 2
	for y := y0; y <= y1; y++ {
		t := float64(y) / half // 0 at the top, 1 at the horizon
		if t > 1 {
			t = 1
		}
		setPixel(fb, cfg.Width, x, y, scaleColor(lerpColor(palette.skyTop, palette.skyHorizon, t), gloom))
	}
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

func sampleFlat(fb []byte, cfg Config, x, y int, rowDist, px, py, dx, dy float64, tex *texture, f float64) {
	wx := px + rowDist*dx
	wy := py + rowDist*dy
	tcx := int(float64(tex.w) * (wx - math.Floor(wx)))
	tcy := int(float64(tex.h) * (wy - math.Floor(wy)))
	setPixel(fb, cfg.Width, x, y, scaleColor(tex.at(tcx, tcy), f))
}

func shadeFactor(dist float64) float64 {
	return max(shadeFloor, min(1, 1.0/(1.0+dist*shadeDecay)))
}

func scaleColor(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: 255,
	}
}
