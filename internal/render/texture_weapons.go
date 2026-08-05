package render

import "image/color"

// This file holds the procedural generators for the first-person weapon
// viewmodels and the muzzle flash. Each draws into the lower-centre of a
// transparent texture; the renderer scales them to the screen.

func genFists() *texture {
	t := newTexture(texSize, texSize)
	flesh := color.RGBA{R: 150, G: 96, B: 74, A: 255}
	edge := adjust(flesh, -40)
	fillRect(t, 6, 44, 26, 64, flesh)
	fillRect(t, 6, 44, 26, 47, edge)
	fillRect(t, 38, 44, 58, 64, flesh)
	fillRect(t, 38, 44, 58, 47, edge)
	return t
}

func genPistol() *texture {
	t := newTexture(texSize, texSize)
	metal := color.RGBA{R: 78, G: 80, B: 92, A: 255}
	dark := color.RGBA{R: 40, G: 40, B: 48, A: 255}
	fillRect(t, 28, 22, 37, 42, metal) // barrel
	fillRect(t, 24, 38, 41, 52, dark)  // body
	fillRect(t, 30, 48, 42, 64, dark)  // grip
	return t
}

func genShotgun() *texture {
	t := newTexture(texSize, texSize)
	metal := color.RGBA{R: 86, G: 88, B: 98, A: 255}
	wood := color.RGBA{R: 110, G: 70, B: 40, A: 255}
	fillRect(t, 20, 18, 44, 40, metal) // twin barrels
	fillRect(t, 31, 18, 33, 40, adjust(metal, -40))
	fillRect(t, 16, 38, 48, 64, wood) // stock/body
	return t
}

func genChaingun() *texture {
	t := newTexture(texSize, texSize)
	metal := color.RGBA{R: 70, G: 72, B: 84, A: 255}
	dark := color.RGBA{R: 40, G: 40, B: 48, A: 255}
	// A cluster of rotating barrels over a chunky body.
	for _, bx := range []int{26, 30, 34} {
		fillRect(t, bx, 16, bx+3, 40, metal)
	}
	fillRect(t, 22, 38, 44, 56, dark)
	fillRect(t, 28, 52, 40, 64, dark) // grip
	return t
}

func genRocketLauncher() *texture {
	t := newTexture(texSize, texSize)
	tube := color.RGBA{R: 80, G: 84, B: 70, A: 255}
	dark := color.RGBA{R: 44, G: 46, B: 40, A: 255}
	fillRect(t, 18, 26, 48, 40, tube) // launch tube
	fillRect(t, 18, 26, 48, 29, adjust(tube, 30))
	fillRect(t, 20, 24, 30, 28, dark) // sight
	fillRect(t, 26, 40, 40, 64, dark) // body/grip
	return t
}

// genFlash draws a bright muzzle-flash blob with a transparent background.
func genFlash() *texture {
	t := newTexture(texSize, texSize)
	cx, cy := 32.0, 32.0
	for y := range texSize {
		for x := range texSize {
			d := ((float64(x)-cx)/16)*((float64(x)-cx)/16) + ((float64(y)-cy)/16)*((float64(y)-cy)/16)
			if d > 1 {
				continue
			}
			if d < 0.5 {
				t.set(x, y, color.RGBA{R: 255, G: 250, B: 200, A: 255})
			} else {
				t.set(x, y, color.RGBA{R: 250, G: 200, B: 70, A: 255})
			}
		}
	}
	return t
}
