package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// maxDDASteps bounds the boundary walk so a ray that somehow escapes the bounded
// world can never loop forever.
const maxDDASteps = 4096

// ceilingDim darkens ceilings slightly relative to floors at the same distance.
const ceilingDim = 0.85

// drawScene renders the level geometry — walls, step faces, floors and ceilings
// — in a single pass over the screen columns, honouring per-tile floor and
// ceiling heights, and records each column's closing distance in zbuf for sprite
// occlusion.
//
// Each column walks its ray boundary by boundary through the grid. A clip window
// [yTop, yBot] tracks the rows still unpainted: at every boundary the surfaces of
// the tile being left are filled up to their projected far edges, any rise in
// floor (or drop in ceiling) across the boundary is drawn as a textured step
// face, and the window tightens. A solid tile paints the remaining window as a
// full wall and closes the column.
func drawScene(fb []byte, zbuf []float64, g *sim.Game, cam camera, cfg Config, tx *textureSet, gloom float64) {
	eyeZ := g.EyeZ()
	for x := range cfg.Width {
		zbuf[x] = drawColumn(fb, g, cam, cfg, tx, x, eyeZ, gloom)
	}
}

// drawColumn renders one screen column and returns the distance at which it
// closed (its occlusion depth). gloom scales all surface lighting.
func drawColumn(fb []byte, g *sim.Game, cam camera, cfg Config, tx *textureSet, x int, eyeZ, gloom float64) float64 {
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

		// Fill the departed tile's floor and ceiling up to this boundary. The
		// spans self-clamp to empty when a surface is out of view (e.g. a floor
		// above eye level, whose step face was drawn at the previous boundary).
		aLight := g.World.Level.LightAt(aX, aY) * gloom
		floorEdge := row(aFloor, d)
		ftex := tx.floor
		if g.World.HazardAt(aX, aY) > 0 {
			ftex = tx.nukage
		}
		fillFloorSpan(fb, cfg, x, max(yTop, floorEdge+1), yBot, aFloor, eyeZ, px, py, dx, dy, ftex, aLight)
		ceilEdge := row(aCeil, d)
		fillCeilSpan(fb, cfg, x, yTop, min(yBot, ceilEdge), aCeil, eyeZ, px, py, dx, dy, tx.ceiling, aLight)

		// The texture column for any face on this boundary, themed by the room
		// it's seen from and lit by the cell it faces.
		texX, tex := boundaryTexture(g, tx, mapX, mapY, side, d, px, py, dx, dy, g.World.Level.ThemeAt(aX, aY))
		bLight := g.World.Level.LightAt(mapX, mapY) * gloom

		bFloor := g.World.FloorAt(mapX, mapY)
		bCeil := g.World.CeilAt(mapX, mapY)
		if g.World.Solid(mapX, mapY) {
			if top := g.World.Level.WallTopAt(mapX, mapY); top > 0 {
				// A low wall: a solid block we can see over. Treat it like an
				// unclimbable step up to its top, then keep walking the ray so the
				// room beyond is drawn above it.
				bFloor, bCeil = top, 1
			} else {
				drawWallSpan(fb, cfg, x, max(yTop, ceilEdge+1), min(yBot, floorEdge), d, eyeZ, texX, tex, side, bLight)
				return d
			}
		}
		if bFloor > aFloor { // rising step face
			drawWallSpan(fb, cfg, x, max(yTop, row(bFloor, d)+1), min(yBot, floorEdge), d, eyeZ, texX, tex, side, bLight)
		}
		if bCeil < aCeil { // dropping ceiling face
			drawWallSpan(fb, cfg, x, max(yTop, ceilEdge+1), min(yBot, row(bCeil, d)), d, eyeZ, texX, tex, side, bLight)
		}

		yBot = min(yBot, row(math.Max(aFloor, bFloor), d))
		yTop = max(yTop, row(math.Min(aCeil, bCeil), d)+1)
		if yTop > yBot {
			return d
		}
		aFloor, aCeil = bFloor, bCeil
		aX, aY = mapX, mapY
	}
	return math.MaxFloat64
}

// boundaryTexture picks the texture and texture column for a face crossed at
// distance d, using where along the cell edge the ray landed (flipped so the
// image faces the camera consistently).
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

// drawWallSpan paints rows y0..y1 of a vertical face at distance d, mapping each
// row to its world height so the texture tiles once per wall unit.
func drawWallSpan(fb []byte, cfg Config, x, y0, y1 int, d, eyeZ float64, texX int, tex *texture, side int, light float64) {
	fh := float64(cfg.Height)
	for y := y0; y <= y1; y++ {
		z := eyeZ - (float64(y)-fh/2)*d/fh
		v := int((1 - z) * float64(tex.h)) // tex.at wraps, tiling tall faces
		setPixel(fb, cfg.Width, x, y, scaleColor(shade(tex.at(texX, v), d, side), light))
	}
}

// fillFloorSpan paints rows y0..y1 of a horizontal floor surface at height z,
// recovering each row's world position from its distance along the ray. light is
// the surface tile's brightness.
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

// fillCeilSpan paints rows y0..y1 of a ceiling surface at height z.
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

// sampleFlat samples a horizontal surface's texture at the world point rowDist
// along this column's ray and writes the shaded pixel.
func sampleFlat(fb []byte, cfg Config, x, y int, rowDist, px, py, dx, dy float64, tex *texture, f float64) {
	wx := px + rowDist*dx
	wy := py + rowDist*dy
	tcx := int(float64(tex.w) * (wx - math.Floor(wx)))
	tcy := int(float64(tex.h) * (wy - math.Floor(wy)))
	setPixel(fb, cfg.Width, x, y, scaleColor(tex.at(tcx, tcy), f))
}

// shadeFactor is the distance-dimming multiplier used for flat surfaces, matching
// the wall shading curve (see shade) for a north/south face.
func shadeFactor(dist float64) float64 {
	return math.Max(shadeFloor, math.Min(1, 1.0/(1.0+dist*shadeDecay)))
}

// scaleColor multiplies an RGB colour by f, keeping it opaque.
func scaleColor(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: 255,
	}
}
