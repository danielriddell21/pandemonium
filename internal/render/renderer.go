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
	loZ         []float64 // per-column nearest occluding lip distance (sprite clip)
	loH         []float64 // per-column world height of that lip
	loRow       []int     // per-column screen row of that lip's top edge
	tex         *textureSet
	overlay     *hud.Overlay
	diagnostics bool
	showMap     bool
	hideHUD     bool    // when set, Frame draws only the world view (no weapon or status bar)
	crosshair   bool    // when set, Frame draws a small aiming cross at view centre
	gloom       float64 // global light multiplier (1 = normal); falls as the run deepens
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
// are off by default; enable them with WithDiagnostics or SetDiagnostics.
func NewRenderer(cfg Config, opts ...Option) *Renderer {
	r := &Renderer{
		cfg:   cfg,
		fb:    make([]byte, cfg.Width*cfg.Height*4),
		zbuf:  make([]float64, cfg.Width),
		loZ:   make([]float64, cfg.Width),
		loH:   make([]float64, cfg.Width),
		loRow: make([]int, cfg.Width),
		tex:   loadTextures(assetDir()),
		gloom: 1,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Config returns the renderer's presentation settings.
func (r *Renderer) Config() Config { return r.cfg }

// SetAutomap toggles whether Frame overlays the explored-level minimap.
func (r *Renderer) SetAutomap(on bool) { r.showMap = on }

// SetHUD toggles the on-screen chrome — the held weapon, the status bar and any
// notices. Turning it off leaves just the rendered world, which suits clean
// screenshots of the scene itself.
func (r *Renderer) SetHUD(on bool) { r.hideHUD = !on }

// SetCrosshair toggles a small aiming cross at the centre of the view.
func (r *Renderer) SetCrosshair(on bool) { r.crosshair = on }

// SetDiagnostics toggles the on-screen diagnostic (playtest) message channel.
func (r *Renderer) SetDiagnostics(on bool) { r.diagnostics = on }

// SetFOV changes the horizontal field of view (radians, clamped to a sane
// range) for subsequent frames.
func (r *Renderer) SetFOV(fov float64) {
	if fov < 0.6 {
		fov = 0.6
	} else if fov > 1.8 {
		fov = 1.8
	}
	r.cfg.FOV = fov
}

// SetGloom sets the global light multiplier (clamped to [0.3, 1]); the app lowers
// it as the run goes deeper so the world darkens toward the end.
func (r *Renderer) SetGloom(g float64) {
	if g < 0.3 {
		g = 0.3
	} else if g > 1 {
		g = 1
	}
	r.gloom = g
}

// Frame renders the current state of g and returns the RGBA buffer (row-major,
// 4 bytes per pixel). The slice is owned by the Renderer and overwritten on the
// next call, so callers should upload or copy it before calling again.
func (r *Renderer) Frame(g *sim.Game) []byte {
	cam := newCamera(g.Player.Angle, r.cfg.FOV)
	drawScene(r.fb, r.zbuf, r.loZ, r.loH, r.loRow, g, cam, r.cfg, r.tex, r.gloom)
	drawSprites(r.fb, r.zbuf, r.loZ, r.loH, r.loRow, g, cam, r.cfg, r.tex)
	drawPowerupTint(r.fb, r.cfg, g)
	if !r.hideHUD {
		weapon := r.tex.weapon[int(g.Player.Weapon)%len(r.tex.weapon)]
		drawViewmodel(r.fb, r.cfg, weapon, r.tex.flash, g.MuzzleFlash(), float64(g.Tick64()), r.cfg.Height-StatusBarH)
		drawStatusBar(r.fb, r.cfg, g, r.tex)
		drawNotice(r.fb, r.cfg, g.Notice())
		if r.overlay != nil {
			if msg, ch, ok := r.overlay.Active(); ok && (ch == hud.Notice || r.diagnostics) {
				drawMessage(r.fb, r.cfg, msg, ch)
			}
		}
	}
	if r.crosshair { // an aiming aid sits on top of the world and the weapon
		drawCrosshair(r.fb, r.cfg, r.hideHUD)
	}
	if r.showMap {
		drawAutomap(r.fb, r.cfg, g)
	}
	return r.fb
}

// Intermission renders the between-levels tally screen for stats and returns the
// RGBA buffer (owned by the Renderer, overwritten on the next call).
func (r *Renderer) Intermission(stats sim.LevelStats) []byte {
	drawIntermission(r.fb, r.cfg, stats)
	return r.fb
}
