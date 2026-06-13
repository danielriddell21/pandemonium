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

// textureSet holds the textures the renderer draws with.
type textureSet struct {
	wall   *texture
	door   *texture
	sprite []*texture
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
	for i := range ts.sprite {
		if t, ok := loadPNG(filepath.Join(dir, fmt.Sprintf("demon%d.png", i))); ok {
			ts.sprite[i] = t
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
		wall: genBrick(palette.wall),
		door: genDoor(palette.door),
		sprite: []*texture{
			genDemon(palette.sprite[0]),
			genDemon(palette.sprite[1]),
		},
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

// genDemon draws a simple demon silhouette with a transparent background.
func genDemon(body color.RGBA) *texture {
	t := newTexture(texSize, texSize)
	cx, cy := 32.0, 38.0
	rx, ry := 20.0, 24.0
	edge := adjust(body, -50)
	eye := color.RGBA{R: 240, G: 220, B: 60, A: 255}
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
	// two eyes
	for _, ex := range []int{25, 39} {
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				t.set(ex+dx, 30+dy, eye)
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
