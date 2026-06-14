package render

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// texSize is the edge length of the procedurally generated textures.
const texSize = 64

// texture is a small RGBA image sampled by the renderer. Alpha 0 marks
// transparent texels (used for sprite cut-outs).
type texture struct {
	w, h int
	pix  []color.RGBA
}

func newTexture(w, h int) *texture {
	return &texture{w: w, h: h, pix: make([]color.RGBA, w*h)}
}

func (t *texture) set(x, y int, c color.RGBA) { t.pix[y*t.w+x] = c }

// at returns the texel at (u, v), wrapping out-of-range coordinates.
func (t *texture) at(u, v int) color.RGBA {
	if t.w == 0 || t.h == 0 {
		return color.RGBA{}
	}
	u %= t.w
	if u < 0 {
		u += t.w
	}
	v %= t.h
	if v < 0 {
		v += t.h
	}
	return t.pix[v*t.w+u]
}

// demonArt holds a demon variant's animation frames.
type demonArt struct {
	walk []*texture // walk cycle
	dead []*texture // death sequence (last frame is the settled corpse)
}

// textureSet holds the textures the renderer draws with.
type textureSet struct {
	wall     *texture
	door     *texture
	demon    []demonArt
	fireball *texture
	weapon   []*texture // indexed by sim.WeaponKind: fists, pistol, shotgun
	flash    *texture   // muzzle flash
}

// loadTextures returns the procedural texture set, overriding any individual
// texture with a PNG found in dir (wall.png, door.png, demon0.png, demon1.png).
// Missing or unreadable files leave the procedural default in place.
func loadTextures(dir string) *textureSet {
	ts := defaultTextures()
	if dir == "" {
		return ts
	}
	if t, ok := loadPNG(filepath.Join(dir, "wall.png")); ok {
		ts.wall = t
	}
	if t, ok := loadPNG(filepath.Join(dir, "door.png")); ok {
		ts.door = t
	}
	for i := range ts.demon {
		if t, ok := loadPNG(filepath.Join(dir, fmt.Sprintf("demon%d.png", i))); ok {
			ts.demon[i].walk = []*texture{t}
		}
	}
	return ts
}

// assetDir is where override textures are looked for.
func assetDir() string {
	if d := os.Getenv("PANDEMONIUM_ASSETS"); d != "" {
		return d
	}
	return "assets/textures"
}

func loadPNG(path string) (*texture, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()
	img, err := png.Decode(f)
	if err != nil {
		return nil, false
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return nil, false
	}
	t := newTexture(w, h)
	for y := range h {
		for x := range w {
			r, g, bl, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			t.set(x, y, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: uint8(a >> 8)})
		}
	}
	return t, true
}

// defaultTextures generates the built-in placeholder textures from the palette.
func defaultTextures() *textureSet {
	return &textureSet{
		wall:     genBrick(palette.wall),
		door:     genDoor(palette.door),
		demon:    []demonArt{buildDemon(palette.sprite[0]), buildDemon(palette.sprite[1])},
		fireball: genFireball(),
		weapon:   []*texture{genFists(), genPistol(), genShotgun()},
		flash:    genFlash(),
	}
}

func fillRect(t *texture, x0, y0, x1, y1 int, c color.RGBA) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if x >= 0 && x < t.w && y >= 0 && y < t.h {
				t.set(x, y, c)
			}
		}
	}
}

// genFists, genPistol, genShotgun draw crude first-person weapon viewmodels into
// the lower-centre of a transparent texture; the renderer scales them to screen.
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

// buildDemon makes a variant's walk and death frames.
func buildDemon(c color.RGBA) demonArt {
	return demonArt{
		walk: []*texture{genDemonWalk(c, 0), genDemonWalk(c, 1)},
		dead: []*texture{genDemonDead(c, 0, 3), genDemonDead(c, 1, 3), genDemonDead(c, 2, 3)},
	}
}

// genBrick draws a brick/mortar pattern with subtle per-brick variation.
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

// genDoor draws a panelled door distinct from the walls.
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

func adjust(c color.RGBA, d int) color.RGBA {
	return color.RGBA{R: clampByte(int(c.R) + d), G: clampByte(int(c.G) + d), B: clampByte(int(c.B) + d), A: c.A}
}

func clampByte(v int) uint8 {
	switch {
	case v < 0:
		return 0
	case v > 255:
		return 255
	default:
		return uint8(v)
	}
}
