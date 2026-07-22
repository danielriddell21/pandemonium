package render

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const texSize = 64

type texture struct {
	w, h int
	pix  []color.RGBA
}

func newTexture(w, h int) *texture {
	return &texture{w: w, h: h, pix: make([]color.RGBA, w*h)}
}

func (t *texture) set(x, y int, c color.RGBA) { t.pix[y*t.w+x] = c }

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

type demonArt struct {
	walk []*texture
	dead []*texture
}

var wallThemes = [world.NumThemes]color.RGBA{
	palette.wall,
	{R: 96, G: 104, B: 130, A: 255},
	{R: 150, G: 78, B: 70, A: 255},
}

type textureSet struct {
	wall      *texture
	walls     [world.NumThemes]*texture
	door      *texture
	floor     *texture
	ceiling   *texture
	nukage    *texture
	lava      *texture
	switchTex *texture
	demon     []demonArt
	fireball  *texture
	rocket    *texture
	barrel    *texture
	weapon    []*texture
	flash     *texture
	face      []*texture
	item      []*texture
}

func (ts *textureSet) itemTexture(k world.ItemKind) *texture {
	if int(k) < len(ts.item) {
		return ts.item[k]
	}
	return nil
}

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

func defaultTextures() *textureSet {
	return &textureSet{
		wall:      genBrick(palette.wall),
		walls:     buildWallThemes(),
		door:      genDoor(palette.door),
		floor:     genFloor(),
		ceiling:   genCeiling(),
		nukage:    genNukage(),
		lava:      genLava(),
		switchTex: genSwitch(),
		demon:     buildDemons(),
		fireball:  genFireball(),
		rocket:    genRocket(),
		barrel:    genBarrel(),
		weapon:    []*texture{genFists(), genPistol(), genShotgun(), genChaingun(), genRocketLauncher()},
		flash:     genFlash(),
		face:      buildFaces(),
		item:      defaultItemTextures(),
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
