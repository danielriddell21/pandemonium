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
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyControlLeft) || inpututil.IsKeyJustPressed(ebiten.KeyF) {
		in.Attack = true
	}

	in.SelectWeapon = g.weaponSelect()
	return in
}

// weaponSelect maps number keys and the mouse wheel to a weapon slot (1=fists,
// 2=pistol, 3=shotgun), or 0 to keep the current weapon. Number keys win.
func (g *Game) weaponSelect() int {
	if _, wy := ebiten.Wheel(); wy != 0 {
		cur := int(g.sim.Player.Weapon)
		if wy > 0 {
			return (cur+1)%3 + 1
		}
		return (cur+2)%3 + 1
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.Key1):
		return 1
	case inpututil.IsKeyJustPressed(ebiten.Key2):
		return 2
	case inpututil.IsKeyJustPressed(ebiten.Key3):
		return 3
	}
	return 0
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
