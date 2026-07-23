package render

import (
	"github.com/danielriddell21/crucible/geom"
	"github.com/danielriddell21/crucible/hud"
	"github.com/danielriddell21/crucible/raycast"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

type Renderer struct {
	cfg         Config
	fb          []byte
	zbuf        []float64
	loZ         []float64
	loH         []float64
	loRow       []int
	tex         *textureSet
	overlay     *hud.Overlay
	diagnostics bool
	showMap     bool
	hideHUD     bool
	crosshair   bool
	gloom       float64
}

type Option func(*Renderer)

func WithOverlay(o *hud.Overlay) Option {
	return func(r *Renderer) { r.overlay = o }
}

func WithDiagnostics(on bool) Option {
	return func(r *Renderer) { r.diagnostics = on }
}

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

func (r *Renderer) Config() Config { return r.cfg }

func (r *Renderer) SetAutomap(on bool) { r.showMap = on }

func (r *Renderer) SetHUD(on bool) { r.hideHUD = !on }

func (r *Renderer) SetCrosshair(on bool) { r.crosshair = on }

func (r *Renderer) SetDiagnostics(on bool) { r.diagnostics = on }

func (r *Renderer) SetFOV(fov float64) {
	if fov < 0.6 {
		fov = 0.6
	} else if fov > 1.8 {
		fov = 1.8
	}
	r.cfg.FOV = fov
}

func (r *Renderer) SetGloom(g float64) {
	if g < 0.3 {
		g = 0.3
	} else if g > 1 {
		g = 1
	}
	r.gloom = g
}

func (r *Renderer) Frame(g *sim.Game) []byte {
	cam := raycast.NewCamera(geom.Vec2{X: g.Player.Pos.X, Y: g.Player.Pos.Y}, g.Player.Angle, r.cfg.FOV)
	drawScene(r.fb, r.zbuf, r.loZ, r.loH, r.loRow, g, cam, r.cfg, r.tex, r.gloom)
	drawSprites(r.fb, r.zbuf, r.loZ, r.loH, r.loRow, g, cam, r.cfg, r.tex)
	drawPowerupTint(r.fb, r.cfg, g)
	if !r.hideHUD {
		weapon := r.tex.weapon[int(g.Player.Weapon)%len(r.tex.weapon)]
		drawViewmodel(r.fb, r.cfg, weapon, r.tex.flash, g.MuzzleFlash(), float64(g.Tick64()), r.cfg.Height-StatusBarH)
		drawStatusBar(r.fb, r.cfg, g, r.tex)
		if notice := g.Notice(); notice != "" {
			drawNotice(r.fb, r.cfg, notice)
		} else {
			drawInteractPrompt(r.fb, r.cfg, g)
		}
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

func (r *Renderer) Intermission(stats sim.LevelStats) []byte {
	drawIntermission(r.fb, r.cfg, stats)
	return r.fb
}
