// Package render is the Ebiten front-end. It reads simulation state and draws a
// first-person view of it with a per-column raycaster; it holds no game logic.
// The dependency direction is strictly render -> sim -> world.
package render

import "image/color"

// Config holds presentation settings independent of game logic.
type Config struct {
	Width  int     // internal render width in pixels
	Height int     // internal render height in pixels
	FOV    float64 // horizontal field of view in radians
}

// DefaultConfig returns sensible defaults: a 640x400 view with a 66° field of
// view, close to the classic Wolfenstein/DOOM feel.
func DefaultConfig() Config {
	return Config{Width: 640, Height: 400, FOV: 1.152}
}

// palette holds the flat colours used when no textures are loaded. They are
// chosen for the dim, brownish DOOM mood and shaded by distance at draw time.
var palette = struct {
	ceiling color.RGBA
	floor   color.RGBA
	wallX   color.RGBA // walls hit on an east/west face
	wallY   color.RGBA // walls hit on a north/south face (drawn darker)
	door    color.RGBA
	sprite  [2]color.RGBA
	hudText color.RGBA // player-facing notice text
	hudDiag color.RGBA // diagnostic/playtest readout text
	hudDrop color.RGBA // message drop shadow
}{
	ceiling: color.RGBA{R: 28, G: 26, B: 30, A: 255},
	floor:   color.RGBA{R: 44, G: 36, B: 30, A: 255},
	wallX:   color.RGBA{R: 150, G: 110, B: 78, A: 255},
	wallY:   color.RGBA{R: 110, G: 80, B: 56, A: 255},
	door:    color.RGBA{R: 120, G: 70, B: 60, A: 255},
	sprite: [2]color.RGBA{
		{R: 168, G: 52, B: 44, A: 255},
		{R: 120, G: 40, B: 96, A: 255},
	},
	hudText: color.RGBA{R: 222, G: 214, B: 188, A: 255},
	hudDiag: color.RGBA{R: 120, G: 200, B: 120, A: 255},
	hudDrop: color.RGBA{R: 0, G: 0, B: 0, A: 255},
}
