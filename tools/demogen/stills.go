package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// stillCfg is the render size for the static feature shots: the same internal
// resolution the animated clips use, so stills and clips match visually.
var stillCfg = render.Config{Width: 256, Height: 160, FOV: 1.152}

// ceilingFill is the dark backdrop used to pad montage cells that hold no frame.
var ceilingFill = color.RGBA{R: 28, G: 26, B: 30, A: 255}

// recordStills renders the static feature shots — bestiary, arsenal, pickups,
// status bar, tally and hazards — that read better as a single frame than as
// motion.
func recordStills() error {
	stills := []struct {
		name string
		draw func(render.Config) image.Image
	}{
		{"bestiary.png", bestiaryShot},
		{"arsenal.png", arsenalShot},
		{"pickups.png", pickupsShot},
		{"hud.png", hudShot},
		{"tally.png", tallyShot},
		{"hazards.png", hazardsShot},
	}
	for _, s := range stills {
		img := s.draw(stillCfg)
		path := filepath.Join(outDir, s.name)
		if err := savePNG(path, img); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
		b := img.Bounds()
		fmt.Printf("wrote %-26s %dx%d still\n", path, b.Dx(), b.Dy())
	}
	return nil
}

// arena builds a flat, walled room with the player at the near wall facing the
// far wall — a clean backdrop for posing demons, items and weapons without any
// generated clutter. Interior cells are open floor (the tile zero value).
func arena(w, h int) *world.Level {
	l := &world.Level{
		Width: w, Height: h,
		Tiles:   make([]world.TileType, w*h),
		FloorH:  make([]float64, w*h),
		CeilH:   make([]float64, w*h),
		Light:   make([]float64, w*h),
		Theme:   make([]uint8, w*h),
		WallTop: make([]float64, w*h),
		Spawn:   world.Coord{X: w / 2, Y: h - 2},
		Exit:    world.Coord{X: w / 2, Y: 1},
	}
	for i := range l.Tiles {
		l.CeilH[i] = 1
		l.Light[i] = 1
	}
	for y := range h {
		for x := range w {
			if x == 0 || y == 0 || x == w-1 || y == h-1 {
				l.Tiles[y*w+x] = world.TileWall
			}
		}
	}
	return l
}

// arenaGame builds an arena simulation stripped of the demons, items and
// projectiles that New scatters, ready for posing a specific feature.
func arenaGame(w, h int) *sim.Game {
	g := sim.New(arena(w, h))
	g.Entities = nil
	g.Items = nil
	g.Projectiles = nil
	return g
}

// frameImage renders one frame of a game and returns it as an owned RGBA image.
// With hud false the chrome (held weapon and status bar) is suppressed, leaving a
// clean view of the world — used for the posed feature shots.
func frameImage(g *sim.Game, cfg render.Config, hud bool) *image.RGBA {
	r := render.NewRenderer(cfg)
	r.SetHUD(hud)
	img := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	copy(img.Pix, r.Frame(g))
	return img
}

// bestiaryShot lines up the five demon kinds across the player's view so each
// distinct silhouette can be compared side by side.
func bestiaryShot(cfg render.Config) image.Image {
	g := arenaGame(18, 14)
	cx, py := g.Player.Pos.X, g.Player.Pos.Y
	kinds := []sim.EntityKind{sim.Melee, sim.Ranged, sim.Gunner, sim.Pinky, sim.Baron}
	offsets := []float64{-2.6, -1.3, 0, 1.3, 2.6}
	for i, k := range kinds {
		g.Entities = append(g.Entities, sim.Entity{
			Pos:    sim.Vec2{X: cx + offsets[i], Y: py - 5},
			Sprite: i,
			Kind:   k,
			State:  sim.Active,
			Health: 100,
			Alive:  true,
		})
	}
	return frameImage(g, cfg, false)
}

// arsenalShot renders each weapon's first-person viewmodel against the same plain
// arena and tiles the cropped viewmodels into a single row.
func arsenalShot(cfg render.Config) image.Image {
	guns := []sim.WeaponKind{sim.Fists, sim.Pistol, sim.Shotgun, sim.Chaingun, sim.RocketLauncher}
	// Crop the bottom-centre region that holds the held weapon (vw = Width*2/5,
	// resting on the status-bar baseline), with a little headroom around it.
	const cropW, cropH = 128, 112
	cropX := cfg.Width/2 - cropW/2
	cropY := cfg.Height - render.StatusBarH - cropH
	montage := image.NewRGBA(image.Rect(0, 0, cropW*len(guns), cropH))
	for i, wk := range guns {
		g := arenaGame(12, 12)
		g.Player.Weapon = wk
		frame := frameImage(g, cfg, true) // the weapon is the subject, so keep it drawn
		dst := image.Rect(i*cropW, 0, (i+1)*cropW, cropH)
		draw.Draw(montage, dst, frame, image.Pt(cropX, cropY), draw.Src)
	}
	return montage
}

// pickupsShot renders each collectible close up in its own cell and tiles them
// into a two-row grid: the common supplies on top, the power-ups below. Item
// billboards are small in the world, so a close-up montage reads far better than
// a single posed frame.
func pickupsShot(cfg render.Config) image.Image {
	supplies := []world.ItemKind{
		world.ItemHealth, world.ItemArmor, world.ItemBullets,
		world.ItemShells, world.ItemRockets, world.ItemBackpack,
	}
	powerups := []world.ItemKind{
		world.ItemSoul, world.ItemMega, world.ItemBerserk,
		world.ItemInvuln, world.ItemRadSuit,
	}
	cols := len(supplies)
	// Crop a window around the close item, stopping just above the status bar so
	// its text does not bleed into the cell.
	const cropW, cropH = 92, 80
	cropX := cfg.Width/2 - cropW/2
	const cropY = 40
	montage := image.NewRGBA(image.Rect(0, 0, cropW*cols, cropH*2))
	draw.Draw(montage, montage.Bounds(), image.NewUniform(ceilingFill), image.Point{}, draw.Src)
	cell := func(k world.ItemKind) *image.RGBA {
		g := arenaGame(10, 10)
		g.Items = []sim.ItemState{{Kind: k, Pos: sim.Vec2{X: g.Player.Pos.X, Y: g.Player.Pos.Y - 2.2}}}
		return frameImage(g, cfg, false)
	}
	for i, k := range supplies {
		dst := image.Rect(i*cropW, 0, (i+1)*cropW, cropH)
		draw.Draw(montage, dst, cell(k), image.Pt(cropX, cropY), draw.Src)
	}
	for i, k := range powerups {
		dst := image.Rect(i*cropW, cropH, (i+1)*cropW, cropH*2)
		draw.Draw(montage, dst, cell(k), image.Pt(cropX, cropY), draw.Src)
	}
	return montage
}

// tallyShot renders the between-levels intermission screen with a representative
// clear — kills, items and secrets as percentages, and time against par.
func tallyShot(cfg render.Config) image.Image {
	stats := sim.LevelStats{
		Kills: 11, KillsTotal: 14,
		Items: 6, ItemsTotal: 8,
		Secrets: 1, SecretsTotal: 2,
		Elapsed: 78.5, Par: 96,
	}
	r := render.NewRenderer(cfg)
	img := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	copy(img.Pix, r.Intermission(stats))
	return img
}

// hudShot poses a populated status bar — health and armour, the current weapon's
// ammo, and the three keycards — and crops it to the bottom panel.
func hudShot(cfg render.Config) image.Image {
	g := arenaGame(12, 12)
	g.Player.Health = 88
	g.Player.Armor = 50
	g.Player.Weapon = sim.Chaingun
	g.Player.Backpack = true
	g.Player.Bullets = 154
	g.Player.Keys = map[world.ItemKind]bool{
		world.ItemKeyRed: true, world.ItemKeyBlue: true, world.ItemKeyYellow: true,
	}
	frame := frameImage(g, cfg, true)
	bar := image.NewRGBA(image.Rect(0, 0, cfg.Width, render.StatusBarH))
	draw.Draw(bar, bar.Bounds(), frame, image.Pt(0, cfg.Height-render.StatusBarH), draw.Src)
	return bar
}

// hazardsShot poses the two damaging floors side by side — radioactive slime and
// molten lava — so they read distinctly. Each is a pool laid across the floor
// just ahead of the camera.
func hazardsShot(cfg render.Config) image.Image {
	pool := func(kind world.HazardKind) *image.RGBA {
		g := arenaGame(14, 12)
		cx, cy := int(g.Player.Pos.X), int(g.Player.Pos.Y)
		hz := map[world.Coord]world.HazardCell{}
		for y := cy - 6; y <= cy-2; y++ {
			for x := cx - 3; x <= cx+3; x++ {
				hz[world.Coord{X: x, Y: y}] = world.HazardCell{Rate: 8, Kind: kind}
			}
		}
		g.World.Level.Hazard = hz
		return frameImage(g, cfg, false)
	}
	// Crop off the bare upper ceiling and tile the two pools left and right.
	const cropY, cropH = 28, 128
	montage := image.NewRGBA(image.Rect(0, 0, cfg.Width*2, cropH))
	src := image.Pt(0, cropY)
	draw.Draw(montage, image.Rect(0, 0, cfg.Width, cropH), pool(world.HazardNukage), src, draw.Src)
	draw.Draw(montage, image.Rect(cfg.Width, 0, cfg.Width*2, cropH), pool(world.HazardLava), src, draw.Src)
	return montage
}

// savePNG writes an image to path as a PNG.
func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return png.Encode(f, img)
}
