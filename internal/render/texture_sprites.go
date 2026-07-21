package render

import (
	"image/color"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// This file holds the procedural generators for the billboarded sprites and HUD
// art: collectibles, demons, projectiles, props and the status-bar faces.

// defaultItemTextures builds the collectible sprites indexed by world.ItemKind.
func defaultItemTextures() []*texture {
	items := make([]*texture, world.ItemKeyYellow+1)
	items[world.ItemHealth] = genMedkit()
	items[world.ItemArmor] = genArmor()
	items[world.ItemBullets] = genAmmoBox(color.RGBA{R: 196, G: 170, B: 60, A: 255})
	items[world.ItemShells] = genAmmoBox(color.RGBA{R: 196, G: 70, B: 50, A: 255})
	items[world.ItemRockets] = genAmmoBox(color.RGBA{R: 120, G: 120, B: 130, A: 255})
	items[world.ItemBackpack] = genBackpack()
	items[world.ItemSoul] = genSphere(color.RGBA{R: 70, G: 110, B: 230, A: 255})
	items[world.ItemMega] = genSphere(color.RGBA{R: 230, G: 200, B: 80, A: 255})
	items[world.ItemBerserk] = genMedkit() // a red medkit-like stim, fitting berserk
	items[world.ItemInvuln] = genSphere(color.RGBA{R: 80, G: 230, B: 120, A: 255})
	items[world.ItemRadSuit] = genArmor()
	items[world.ItemKeyRed] = genKey(color.RGBA{R: 210, G: 50, B: 50, A: 255})
	items[world.ItemKeyBlue] = genKey(color.RGBA{R: 70, G: 110, B: 220, A: 255})
	items[world.ItemKeyYellow] = genKey(color.RGBA{R: 220, G: 200, B: 60, A: 255})
	return items
}

// genMedkit draws a white box with a red cross on a transparent background.
func genMedkit() *texture {
	t := newTexture(texSize, texSize)
	box := color.RGBA{R: 230, G: 230, B: 224, A: 255}
	fillRect(t, 18, 22, 46, 50, adjust(box, -40))
	fillRect(t, 19, 23, 45, 49, box)
	red := color.RGBA{R: 200, G: 40, B: 40, A: 255}
	fillRect(t, 29, 28, 35, 44, red) // vertical bar
	fillRect(t, 24, 33, 40, 39, red) // horizontal bar
	return t
}

// genArmor draws a simple green chest-plate on a transparent background.
func genArmor() *texture {
	t := newTexture(texSize, texSize)
	green := color.RGBA{R: 60, G: 170, B: 70, A: 255}
	for y := 22; y < 50; y++ {
		// Taper the plate toward the bottom for a vest-like silhouette.
		inset := (y - 22) / 4
		fillRect(t, 20+inset, y, 44-inset, y+1, green)
	}
	fillRect(t, 20, 22, 44, 25, adjust(green, 40)) // collar highlight
	return t
}

// genAmmoBox draws a small ammo container tinted by the round it holds.
func genAmmoBox(c color.RGBA) *texture {
	t := newTexture(texSize, texSize)
	fillRect(t, 20, 30, 44, 46, adjust(c, -50))
	fillRect(t, 21, 31, 43, 45, c)
	fillRect(t, 21, 31, 43, 34, adjust(c, 40)) // lid highlight
	return t
}

// genKey draws a keycard in the given colour on a transparent background.
func genKey(c color.RGBA) *texture {
	t := newTexture(texSize, texSize)
	fillRect(t, 26, 22, 38, 48, adjust(c, -50))
	fillRect(t, 27, 23, 37, 47, c)
	fillRect(t, 29, 25, 35, 30, adjust(c, 60)) // notch detail
	return t
}

// genSphere draws a glowing orb (soulsphere/megasphere/invulnerability) with a
// bright core fading to the given hue, on a transparent background.
func genSphere(hue color.RGBA) *texture {
	t := newTexture(texSize, texSize)
	cx, cy := 32.0, 34.0
	for y := range texSize {
		for x := range texSize {
			nx, ny := (float64(x)-cx)/16, (float64(y)-cy)/16
			d := nx*nx + ny*ny
			if d > 1 {
				continue
			}
			if d < 0.3 {
				t.set(x, y, color.RGBA{R: 245, G: 245, B: 250, A: 255}) // hot core
			} else {
				t.set(x, y, adjust(hue, int(-40*d)))
			}
		}
	}
	return t
}

// genBackpack draws a brown satchel with straps on a transparent background.
func genBackpack() *texture {
	t := newTexture(texSize, texSize)
	canvas := color.RGBA{R: 120, G: 86, B: 50, A: 255}
	strap := adjust(canvas, -45)
	fillRect(t, 20, 24, 44, 48, canvas)
	fillRect(t, 24, 24, 27, 48, strap)
	fillRect(t, 37, 24, 40, 48, strap)
	fillRect(t, 20, 30, 44, 33, strap) // buckle line
	return t
}

// genRocket draws the player's in-flight rocket: a metal slug with a flame tail.
func genRocket() *texture {
	t := newTexture(texSize, texSize)
	body := color.RGBA{R: 150, G: 150, B: 160, A: 255}
	flame := color.RGBA{R: 250, G: 180, B: 60, A: 255}
	fillRect(t, 26, 22, 38, 44, body)             // casing
	fillRect(t, 28, 18, 36, 24, adjust(body, 30)) // nose
	fillRect(t, 28, 44, 36, 52, flame)            // exhaust
	return t
}

// buildDemons makes the animation sets for every demon variant, one per palette
// sprite colour (melee, ranged, gunner, pinky, baron).
func buildDemons() []demonArt {
	arts := make([]demonArt, len(palette.sprite))
	for i, c := range palette.sprite {
		arts[i] = buildDemon(c)
	}
	return arts
}

// buildDemon makes a variant's walk and death frames.
func buildDemon(c color.RGBA) demonArt {
	return demonArt{
		walk: []*texture{genDemonWalk(c, 0), genDemonWalk(c, 1)},
		dead: []*texture{genDemonDead(c, 0, 3), genDemonDead(c, 1, 3), genDemonDead(c, 2, 3)},
	}
}

// drawBody paints an elliptical demon body (transparent outside) with an edge
// shade and, optionally, two eyes.
func drawBody(t *texture, body color.RGBA, cx, cy, rx, ry float64, eyes bool) {
	edge := adjust(body, -50)
	for y := range texSize {
		for x := range texSize {
			nx := (float64(x) - cx) / rx
			ny := (float64(y) - cy) / ry
			d := nx*nx + ny*ny
			if d > 1 {
				continue // transparent
			}
			c := body
			if d > 0.78 {
				c = edge
			}
			t.set(x, y, c)
		}
	}
	if eyes {
		eye := color.RGBA{R: 240, G: 220, B: 60, A: 255}
		ey := int(cy - 8)
		for _, ex := range []int{int(cx - 7), int(cx + 7)} {
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					if ex+dx >= 0 && ex+dx < texSize && ey+dy >= 0 && ey+dy < texSize {
						t.set(ex+dx, ey+dy, eye)
					}
				}
			}
		}
	}
}

// genDemonWalk draws one walk-cycle frame (step 0 or 1) with a slight bob and
// swapping legs.
func genDemonWalk(body color.RGBA, step int) *texture {
	t := newTexture(texSize, texSize)
	drawBody(t, body, 32, 38+float64(step)*2, 20, 24, true)
	leg := adjust(body, -40)
	for _, bx := range []int{22 + step*6, 42 - step*6} {
		for y := 58; y < 64; y++ {
			for x := bx; x < bx+4 && x < texSize; x++ {
				if x >= 0 {
					t.set(x, y, leg)
				}
			}
		}
	}
	return t
}

// genDemonDead draws death frame k of n: the body squashes toward the floor,
// darkens, and loses its eyes.
func genDemonDead(body color.RGBA, k, n int) *texture {
	t := newTexture(texSize, texSize)
	prog := float64(k) / float64(n-1)
	dark := adjust(body, -int(60*prog))
	ry := 24.0 * (1 - 0.75*prog)
	drawBody(t, dark, 32, 56-ry, 22, ry, k == 0)
	return t
}

// genBarrel draws a stout metal barrel with banding, on a transparent
// background, sitting in the lower-centre so it reads as a short floor prop.
func genBarrel() *texture {
	t := newTexture(texSize, texSize)
	metal := color.RGBA{R: 120, G: 96, B: 48, A: 255}
	band := color.RGBA{R: 70, G: 56, B: 28, A: 255}
	const x0, x1, y0, y1 = 20, 44, 18, 60
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			var c color.RGBA
			switch {
			case x < x0+2 || x >= x1-2: // edge shading for a rounded look
				c = adjust(metal, -40)
			case y%14 < 2: // hoops around the barrel
				c = band
			default:
				c = adjust(metal, ((x+y)%5-2)*4)
			}
			t.set(x, y, c)
		}
	}
	return t
}

// genFireball draws a glowing projectile with a transparent background.
func genFireball() *texture {
	t := newTexture(texSize, texSize)
	cx, cy := 32.0, 32.0
	for y := range texSize {
		for x := range texSize {
			nx := (float64(x) - cx) / 14
			ny := (float64(y) - cy) / 14
			d := nx*nx + ny*ny
			if d > 1 {
				continue
			}
			switch {
			case d < 0.3:
				t.set(x, y, color.RGBA{R: 255, G: 240, B: 180, A: 255})
			case d < 0.7:
				t.set(x, y, color.RGBA{R: 250, G: 150, B: 40, A: 255})
			default:
				t.set(x, y, color.RGBA{R: 200, G: 50, B: 20, A: 255})
			}
		}
	}
	return t
}

// buildFaces makes the status-bar mugshots: one per health band (0 healthy ..
// 3 dead) and gaze direction (-1 left, 0 ahead, +1 right), indexed band*3+dir+1.
func buildFaces() []*texture {
	const bands, dirs = 4, 3
	faces := make([]*texture, bands*dirs)
	for b := range bands {
		for d := -1; d <= 1; d++ {
			faces[b*dirs+d+1] = genFace(b, d)
		}
	}
	return faces
}

// faceIndex maps a health band and gaze direction to its buildFaces slot.
func faceIndex(band, dir int) int { return band*3 + dir + 1 }

// genFace draws the status-bar mugshot for a health band and gaze direction: a
// skin disc with eyes (shifted by gaze), a brow that lowers and a mouth that
// turns from a faint smile to a pained grimace as the band rises. The background
// stays transparent so the bar panel shows through.
func genFace(band, gaze int) *texture {
	t := newTexture(texSize, texSize)
	skins := []color.RGBA{
		{R: 210, G: 162, B: 120, A: 255},
		{R: 204, G: 150, B: 106, A: 255},
		{R: 188, G: 134, B: 96, A: 255},
		{R: 150, G: 124, B: 112, A: 255}, // ashen
	}
	skin := skins[band%len(skins)]
	cx, cy, r := 32.0, 32.0, 22.0
	for y := range texSize {
		for x := range texSize {
			dx, dy := (float64(x)-cx)/r, (float64(y)-cy)/r
			d := dx*dx + dy*dy
			if d > 1 {
				continue
			}
			c := skin
			if d > 0.8 {
				c = adjust(skin, -35)
			}
			t.set(x, y, c)
		}
	}
	eye := color.RGBA{R: 30, G: 20, B: 20, A: 255}
	ex := gaze * 4 // shift the eyes toward where the damage came from
	fillRect(t, 22+ex, 26, 27+ex, 31, eye)
	fillRect(t, 37+ex, 26, 42+ex, 31, eye)
	brow := adjust(skin, -80)
	by := 22 + band*2 // brow lowers (angrier/pained) as health drops
	fillRect(t, 20, by, 44, by+2, brow)
	mouth := color.RGBA{R: 120, G: 40, B: 40, A: 255}
	switch band {
	case 0:
		fillRect(t, 26, 43, 38, 45, mouth)
	case 1:
		fillRect(t, 26, 42, 38, 45, mouth)
	case 2:
		fillRect(t, 25, 41, 39, 46, mouth)
	default:
		fillRect(t, 24, 40, 40, 49, mouth) // open, pained
	}
	return t
}
