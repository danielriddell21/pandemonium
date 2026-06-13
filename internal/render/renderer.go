package render

import (
	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/sim"
)

// Renderer rasterises a first-person view of a simulation into an RGBA pixel
// buffer. It reuses its buffers across frames and depends on nothing but sim and
// world, so it can be driven by any front-end (or none, in tests).
type Renderer struct {
	cfg         Config
	fb          []byte
	zbuf        []float64
	tex         *textureSet
	overlay     *hud.Overlay
	diagnostics bool
}

// Option configures a Renderer.
type Option func(*Renderer)

// WithOverlay attaches a HUD overlay whose active status message is drawn over
// the view.
func WithOverlay(o *hud.Overlay) Option {
	return func(r *Renderer) { r.overlay = o }
}

// WithDiagnostics forces the diagnostic message channel on or off, overriding the
// environment default.
func WithDiagnostics(on bool) Option {
	return func(r *Renderer) { r.diagnostics = on }
}

// NewRenderer builds a renderer for the given configuration. Diagnostic messages
// default to the PANDEMONIUM_DEBUG environment setting.
func NewRenderer(cfg Config, opts ...Option) *Renderer {
	r := &Renderer{
		cfg:         cfg,
		fb:          make([]byte, cfg.Width*cfg.Height*4),
		zbuf:        make([]float64, cfg.Width),
		tex:         loadTextures(assetDir()),
		diagnostics: diagnosticsFromEnv(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
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
	drawHealthBar(r.fb, r.cfg, g.Player.Health/sim.MaxHealth)
	if r.overlay != nil {
		if msg, ch, ok := r.overlay.Active(); ok && (ch == hud.Notice || r.diagnostics) {
			drawMessage(r.fb, r.cfg, msg, ch)
		}
	}
	return r.fb
}
