package render

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/danielriddell21/pandemonium/internal/world"
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
	floor    *texture
	ceiling  *texture
	demon    []demonArt
	fireball *texture
	weapon   []*texture // indexed by sim.WeaponKind: fists, pistol, shotgun
	flash    *texture   // muzzle flash
	face     []*texture // status-bar face, by health band (0 healthy .. 3 dead)
	item     []*texture // indexed by world.ItemKind
}

// itemTexture returns the sprite for a collectible kind.
func (ts *textureSet) itemTexture(k world.ItemKind) *texture {
	if int(k) < len(ts.item) {
		return ts.item[k]
	}
	return nil
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
	if t, ok := loadPNG(filepath.Join(dir, "floor.png")); ok {
		ts.floor = t
	}
	if t, ok := loadPNG(filepath.Join(dir, "ceiling.png")); ok {
		ts.ceiling = t
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
		floor:    genFloor(),
		ceiling:  genCeiling(),
		demon:    []demonArt{buildDemon(palette.sprite[0]), buildDemon(palette.sprite[1]), buildDemon(palette.sprite[2])},
		fireball: genFireball(),
		weapon:   []*texture{genFists(), genPistol(), genShotgun()},
		flash:    genFlash(),
		face:     []*texture{genFace(0), genFace(1), genFace(2), genFace(3)},
		item:     defaultItemTextures(),
	}
}

// defaultItemTextures builds the collectible sprites indexed by world.ItemKind.
func defaultItemTextures() []*texture {
	items := make([]*texture, world.ItemKeyYellow+1)
	items[world.ItemHealth] = genMedkit()
	items[world.ItemArmor] = genArmor()
	items[world.ItemBullets] = genAmmoBox(color.RGBA{R: 196, G: 170, B: 60, A: 255})
	items[world.ItemShells] = genAmmoBox(color.RGBA{R: 196, G: 70, B: 50, A: 255})
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

// genFloor draws a flagstone tile: stone slabs separated by darker grout, with a
// deterministic speckle so the cast floor reads as textured rather than flat.
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

// genCeiling draws a dim, mottled ceiling distinct from the floor so up and down
// read differently once they are cast.
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

// genFace draws the status-bar mugshot for a health band (0 = healthy, 3 = dead):
// a skin disc with eyes, a brow that lowers and a mouth that turns from a faint
// smile to a pained grimace as the band rises. The background stays transparent so
// the bar panel shows through.
func genFace(band int) *texture {
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
	fillRect(t, 22, 26, 27, 31, eye)
	fillRect(t, 37, 26, 42, 31, eye)
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
