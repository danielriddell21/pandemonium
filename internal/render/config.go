package render

import "image/color"

type Config struct {
	Width  int
	Height int
	FOV    float64
}

func DefaultConfig() Config {
	return Config{Width: 640, Height: 400, FOV: 1.152}
}

var palette = struct {
	ceiling    color.RGBA
	floor      color.RGBA
	wall       color.RGBA
	door       color.RGBA
	sprite     [5]color.RGBA
	hudText    color.RGBA
	hudDiag    color.RGBA
	hudDrop    color.RGBA
	skyTop     color.RGBA
	skyHorizon color.RGBA
}{
	ceiling: color.RGBA{R: 28, G: 26, B: 30, A: 255},
	floor:   color.RGBA{R: 44, G: 36, B: 30, A: 255},
	wall:    color.RGBA{R: 150, G: 110, B: 78, A: 255},
	door:    color.RGBA{R: 120, G: 70, B: 60, A: 255},
	sprite: [5]color.RGBA{
		{R: 168, G: 52, B: 44, A: 255},
		{R: 120, G: 40, B: 96, A: 255},
		{R: 96, G: 120, B: 60, A: 255},
		{R: 220, G: 120, B: 150, A: 255},
		{R: 150, G: 110, B: 40, A: 255},
	},
	hudText:    color.RGBA{R: 222, G: 214, B: 188, A: 255},
	hudDiag:    color.RGBA{R: 120, G: 200, B: 120, A: 255},
	hudDrop:    color.RGBA{R: 0, G: 0, B: 0, A: 255},
	skyTop:     color.RGBA{R: 34, G: 40, B: 58, A: 255},
	skyHorizon: color.RGBA{R: 120, G: 96, B: 92, A: 255},
}
