// Package app is the Ebiten application: it owns the window, the game loop and
// input, and draws each frame using the render package. It is the only place
// that imports Ebiten. Dependency direction: app -> render -> sim -> world.
package app

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/sim"
)

// NextFunc produces the simulation for the next level. The app calls it when the
// player reaches an exit, so level progression and per-level wiring stay outside
// the app.
type NextFunc func() *sim.Game

// Game implements ebiten.Game, driving the current simulation and rendering it.
type Game struct {
	sim      *sim.Game
	renderer *render.Renderer
	next     NextFunc

	haveMouse  bool
	lastMouseX int
}

var _ ebiten.Game = (*Game)(nil)

// New builds the application around an initial simulation. next advances to a
// fresh level when the player reaches an exit.
func New(g *sim.Game, renderer *render.Renderer, next NextFunc) *Game {
	return &Game{sim: g, renderer: renderer, next: next}
}

// Update advances the simulation by one tick and swaps in the next level once
// the player reaches an exit.
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	dt := 1.0 / float64(ebiten.TPS())
	g.sim.Tick(g.readInput(), dt)

	if g.sim.ReachedExit() {
		if ng := g.next(); ng != nil {
			g.sim = ng
		}
	}
	return nil
}

// Draw renders the current view and uploads it to the screen.
func (g *Game) Draw(screen *ebiten.Image) {
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
	return ebiten.RunGame(g)
}
