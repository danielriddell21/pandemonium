package render

import (
	"image/color"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func buildWallThemes() [world.NumThemes]*texture {
	var ws [world.NumThemes]*texture
	for i, tint := range wallThemes {
		ws[i] = genBrick(tint)
	}
	return ws
}

func genBrick(base color.RGBA) *texture {
	t := newTexture(texSize, texSize)
	mortar := color.RGBA{R: 58, G: 44, B: 34, A: 255}
	const rowH, brickW, gap = 16, 32, 2
	for y := range texSize {
		row := y / rowH
		offset := (row % 2) * (brickW / 2)
		for x := range texSize {
			bx := (x + offset) % brickW
			by := y % rowH
			if by < gap || bx < gap {
				t.set(x, y, mortar)
				continue
			}
			n := ((x*7 + y*13) % 11) - 5 // deterministic speckle
			t.set(x, y, adjust(base, n*2))
		}
	}
	return t
}

func genDoor(base color.RGBA) *texture {
	t := newTexture(texSize, texSize)
	frame := color.RGBA{R: 70, G: 40, B: 36, A: 255}
	for y := range texSize {
		for x := range texSize {
			var c color.RGBA
			switch {
			case x < 3 || x >= texSize-3 || y < 3 || y >= texSize-3:
				c = frame
			case x%16 < 2: // vertical seams
				c = adjust(base, -30)
			default:
				c = adjust(base, ((x+y)%7-3)*2)
			}
			t.set(x, y, c)
		}
	}
	return t
}

func genFloor() *texture {
	t := newTexture(texSize, texSize)
	base := color.RGBA{R: 78, G: 66, B: 52, A: 255}
	grout := color.RGBA{R: 34, G: 28, B: 22, A: 255}
	const half = texSize / 2
	for y := range texSize {
		for x := range texSize {
			// Two slabs per axis with a grout border around each.
			gx, gy := x%half, y%half
			if gx < 2 || gy < 2 {
				t.set(x, y, grout)
				continue
			}
			n := ((x*13 + y*7) % 11) - 5 // deterministic speckle
			t.set(x, y, adjust(base, n*2))
		}
	}
	return t
}

func genCeiling() *texture {
	t := newTexture(texSize, texSize)
	base := color.RGBA{R: 44, G: 44, B: 56, A: 255}
	for y := range texSize {
		for x := range texSize {
			n := ((x*5 + y*11) % 9) - 4 // deterministic speckle
			t.set(x, y, adjust(base, n*3))
		}
	}
	return t
}

func genSwitch() *texture {
	t := genBrick(palette.wall)
	panel := color.RGBA{R: 40, G: 44, B: 52, A: 255}
	lever := color.RGBA{R: 90, G: 220, B: 120, A: 255}
	fillRect(t, 24, 18, 40, 46, panel)
	fillRect(t, 26, 20, 38, 44, adjust(panel, 20))
	fillRect(t, 30, 22, 34, 40, lever) // the lit lever
	return t
}

func genNukage() *texture {
	t := newTexture(texSize, texSize)
	base := color.RGBA{R: 60, G: 120, B: 40, A: 255}
	for y := range texSize {
		for x := range texSize {
			n := ((x*9 + y*5) % 13) - 6 // coarse, blotchy variation
			t.set(x, y, adjust(base, n*4))
		}
	}
	return t
}

func genLava() *texture {
	t := newTexture(texSize, texSize)
	base := color.RGBA{R: 150, G: 48, B: 20, A: 255}
	for y := range texSize {
		for x := range texSize {
			n := ((x*7 + y*3) % 11) - 5
			c := adjust(base, n*5)
			if (x*x+y*y*3)%17 < 3 { // sparse bright cracks
				c = color.RGBA{R: 240, G: 170, B: 60, A: 255}
			}
			t.set(x, y, c)
		}
	}
	return t
}
