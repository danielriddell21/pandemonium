package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

// mouseSensitivity converts horizontal cursor motion (pixels) into turn radians.
const mouseSensitivity = 0.0035

// readInput samples keyboard and mouse into a simulation input for this tick:
// WASD/arrows move and turn, the mouse turns, and E or Space interacts.
func (g *Game) readInput() sim.Input {
	var in sim.Input

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		in.Forward++
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		in.Forward--
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		in.Strafe++
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		in.Strafe--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		in.Turn++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		in.Turn--
	}

	in.TurnDelta = g.mouseTurn()

	if inpututil.IsKeyJustPressed(ebiten.KeyE) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		in.Interact = true
	}
	return in
}

// mouseTurn returns the turn delta from horizontal cursor movement since the
// last frame, ignoring the first frame's jump.
func (g *Game) mouseTurn() float64 {
	mx, _ := ebiten.CursorPosition()
	if !g.haveMouse {
		g.haveMouse = true
		g.lastMouseX = mx
		return 0
	}
	delta := mx - g.lastMouseX
	g.lastMouseX = mx
	return float64(delta) * mouseSensitivity
}
