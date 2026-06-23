// Package app is the Ebiten application: it owns the window, the game loop and
// input, and draws each frame using the render package. It is the only place
// that imports Ebiten. Dependency direction: app -> render -> sim -> world.
package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/sim"
)

// state is the app's top-level mode: playing a level or showing the tally screen
// between levels.
type state int

const (
	statePlaying state = iota
	stateIntermission
)

// NextFunc produces the simulation for the next level. The app calls it when the
// player reaches an exit, so level progression and per-level wiring stay outside
// the app.
type NextFunc func() *sim.Game

// Game implements ebiten.Game, driving the current simulation and rendering it.
type Game struct {
	sim      *sim.Game
	renderer *render.Renderer
	overlay  *hud.Overlay
	next     NextFunc

	audio     *Audio
	wasFiring bool // muzzle-flash state last tick, for one shot-sound per shot

	state      state
	tally      sim.LevelStats // captured stats shown on the intermission screen
	showMap    bool           // automap overlay toggled with Tab
	depth      int            // levels advanced into the run, dimming the world
	haveMouse  bool
	lastMouseX int
}

var _ ebiten.Game = (*Game)(nil)

// Option configures a Game.
type Option func(*Game)

// WithOverlay attaches a HUD overlay that the loop advances each frame.
func WithOverlay(o *hud.Overlay) Option {
	return func(g *Game) { g.overlay = o }
}

// WithAudio attaches the sound engine. A nil engine leaves the game silent.
func WithAudio(a *Audio) Option {
	return func(g *Game) { g.audio = a }
}

// New builds the application around an initial simulation. next advances to a
// fresh level when the player reaches an exit.
func New(g *sim.Game, renderer *render.Renderer, next NextFunc, opts ...Option) *Game {
	game := &Game{sim: g, renderer: renderer, next: next}
	for _, opt := range opts {
		opt(game)
	}
	return game
}

// Update advances the simulation by one tick. Reaching an exit pauses play on a
// tally screen; pressing Enter there advances to the next level.
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.showMap = !g.showMap
		g.renderer.SetAutomap(g.showMap)
	}

	if g.state == stateIntermission {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			if ng := g.next(); ng != nil {
				g.sim = ng
			}
			g.depth++
			g.renderer.SetGloom(1 - 0.04*float64(g.depth)) // the world darkens as the run deepens
			g.state = statePlaying
		}
		return nil
	}

	dt := 1.0 / float64(ebiten.TPS())
	if g.audio != nil {
		g.audio.SetListener(g.sim.Player.Pos) // attenuate the tick's sounds by distance
	}
	g.sim.Tick(g.readInput(), dt)

	// Play the weapon sound once per shot, on the muzzle-flash rising edge.
	firing := g.sim.MuzzleFlash()
	if firing && !g.wasFiring && g.audio != nil {
		g.audio.Fire()
	}
	g.wasFiring = firing

	if g.overlay != nil {
		g.overlay.Tick()
	}

	if g.sim.LevelComplete() {
		g.tally = g.sim.LevelStats()
		g.state = stateIntermission
	}
	return nil
}

// Draw renders the current view, or the tally screen between levels, and uploads
// it to the screen.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.state == stateIntermission {
		screen.WritePixels(g.renderer.Intermission(g.tally))
		return
	}
	screen.WritePixels(g.renderer.Frame(g.sim))
}

// Layout fixes the internal resolution; Ebiten scales it to the window.
func (g *Game) Layout(_, _ int) (int, int) {
	cfg := g.renderer.Config()
	return cfg.Width, cfg.Height
}

// windowScale enlarges the internal resolution to a comfortable window size.
const windowScale = 2

// Run opens the window and runs the game loop until the player quits. It blocks.
func (g *Game) Run() error {
	cfg := g.renderer.Config()
	ebiten.SetWindowSize(cfg.Width*windowScale, cfg.Height*windowScale)
	ebiten.SetWindowTitle("pandemonium")
	g.audio.StartAmbient()
	return ebiten.RunGame(g)
}
