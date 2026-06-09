package render

import "github.com/danielriddell21/pandemonium/internal/sim"

// Renderer rasterises a first-person view of a simulation into an RGBA pixel
// buffer. It reuses its buffers across frames and depends on nothing but sim and
// world, so it can be driven by any front-end (or none, in tests).
type Renderer struct {
	cfg  Config
	fb   []byte
	zbuf []float64
}

// NewRenderer builds a renderer for the given configuration.
func NewRenderer(cfg Config) *Renderer {
	return &Renderer{
		cfg:  cfg,
		fb:   make([]byte, cfg.Width*cfg.Height*4),
		zbuf: make([]float64, cfg.Width),
	}
}

// Config returns the renderer's presentation settings.
func (r *Renderer) Config() Config { return r.cfg }

// Frame renders the current state of g and returns the RGBA buffer (row-major,
// 4 bytes per pixel). The slice is owned by the Renderer and overwritten on the
// next call, so callers should upload or copy it before calling again.
func (r *Renderer) Frame(g *sim.Game) []byte {
	cam := newCamera(g.Player.Angle, r.cfg.FOV)
	clearBackground(r.fb, r.cfg)
	drawWalls(r.fb, r.zbuf, g, cam, r.cfg)
	drawSprites(r.fb, r.zbuf, g, cam, r.cfg)
	return r.fb
}
