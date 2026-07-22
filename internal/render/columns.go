package render

import (
	"image/color"
	"math"

	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/paint"
	"github.com/danielriddell21/crucible/raycast"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

const ceilingDim = 0.85

// worldHeights adapts the simulation's terrain to the raycaster's Heights
// interface. Floors and ceilings come from the sim so lifts render at their
// current height, not their resting one.
type worldHeights struct{ w *sim.World }

func (h worldHeights) Floor(x, y int) float64   { return h.w.FloorAt(x, y) }
func (h worldHeights) Ceil(x, y int) float64    { return h.w.CeilAt(x, y) }
func (h worldHeights) Solid(x, y int) bool      { return h.w.Solid(x, y) }
func (h worldHeights) WallTop(x, y int) float64 { return h.w.Level.WallTopAt(x, y) }

func drawScene(fb []byte, zbuf, loZ, loH []float64, loRow []int, g *sim.Game, cam raycast.Camera, cfg Config, tx *textureSet, gloom float64) {
	eyeZ := g.EyeZ()
	p := &columnPainter{fb: fb, g: g, cfg: cfg, tx: tx, cam: cam, eyeZ: eyeZ, gloom: gloom}
	hg := worldHeights{w: g.World}
	for x := range cfg.Width {
		res := raycast.WalkColumn(cam, cfg.Width, cfg.Height, x, eyeZ, hg, p)
		zbuf[x], loZ[x], loH[x], loRow[x] = res.Depth, res.LowZ, res.LowH, res.LowRow
	}
}

// columnPainter paints the spans WalkColumn uncovers, keeping pandemonium's
// textures, room themes, hazard floors, open sky, and gloom shading.
type columnPainter struct {
	fb    []byte
	g     *sim.Game
	cfg   Config
	tx    *textureSet
	cam   raycast.Camera
	eyeZ  float64
	gloom float64
}

func (p *columnPainter) WallSpan(x, y0, y1 int, face raycast.Face) {
	light := p.g.World.Level.LightAt(face.Cell.X, face.Cell.Y) * p.gloom
	theme := p.g.World.Level.ThemeAt(face.From.X, face.From.Y) // seen-from room theme
	tex := p.tx.walls[int(theme)%len(p.tx.walls)]
	switch p.g.World.Level.At(face.Cell.X, face.Cell.Y) {
	case world.TileDoor:
		tex = p.tx.door
	case world.TileSwitch:
		tex = p.tx.switchTex
	}
	texX := int(face.WallX * float64(tex.w))
	if texX >= tex.w {
		texX = tex.w - 1
	}
	if face.Flip {
		texX = tex.w - 1 - texX
	}
	drawWallSpan(p.fb, p.cfg, x, y0, y1, face.Dist, p.eyeZ, texX, tex, face.Side, light)
}

func (p *columnPainter) FloorSpan(x, y0, y1 int, cell geom.Coord, z, _ float64) {
	rayX, rayY := p.cam.RayDir(x, p.cfg.Width)
	light := p.g.World.Level.LightAt(cell.X, cell.Y) * p.gloom
	ftex := p.tx.floor
	if p.g.World.HazardAt(cell.X, cell.Y) > 0 {
		ftex = p.tx.nukage
		if p.g.World.Level.HazardKindAt(cell.X, cell.Y) == world.HazardLava {
			ftex = p.tx.lava
		}
	}
	fillFloorSpan(p.fb, p.cfg, x, y0, y1, z, p.eyeZ, p.g.Player.Pos.X, p.g.Player.Pos.Y, rayX, rayY, ftex, light)
}

func (p *columnPainter) CeilSpan(x, y0, y1 int, cell geom.Coord, z, _ float64) {
	if p.g.World.Level.SkyAt(cell.X, cell.Y) {
		// Open air overhead: a bright, distance-independent sky rather than
		// distance-shaded stone (still dimming with the run's gloom).
		fillSkySpan(p.fb, p.cfg, x, y0, y1, p.gloom)
		return
	}
	rayX, rayY := p.cam.RayDir(x, p.cfg.Width)
	light := p.g.World.Level.LightAt(cell.X, cell.Y) * p.gloom
	fillCeilSpan(p.fb, p.cfg, x, y0, y1, z, p.eyeZ, p.g.Player.Pos.X, p.g.Player.Pos.Y, rayX, rayY, p.tx.ceiling, light)
}

func drawWallSpan(fb []byte, cfg Config, x, y0, y1 int, d, eyeZ float64, texX int, tex *texture, side int, light float64) {
	fh := float64(cfg.Height)
	for y := y0; y <= y1; y++ {
		z := eyeZ - (float64(y)-fh/2)*d/fh
		v := int((1 - z) * float64(tex.h)) // tex.at wraps, tiling tall faces
		setPixel(fb, cfg.Width, x, y, paint.Scale(shade(tex.at(texX, v), d, side), light))
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
		setPixel(fb, cfg.Width, x, y, paint.Scale(lerpColor(palette.skyTop, palette.skyHorizon, t), gloom))
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
	setPixel(fb, cfg.Width, x, y, paint.Scale(tex.at(tcx, tcy), f))
}

func shadeFactor(dist float64) float64 {
	return max(shadeFloor, min(1, 1.0/(1.0+dist*shadeDecay)))
}
