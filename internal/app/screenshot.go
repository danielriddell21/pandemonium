package app

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

// activeWorld returns the simulation currently being drawn as a world view, or
// nil when a menu or the tally screen is showing (nothing to photograph).
func (g *Game) activeWorld() *sim.Game {
	switch g.state {
	case statePlaying:
		return g.sim
	case stateAttract:
		return g.attract
	}
	return nil
}

// screenshot renders the current world view without the HUD chrome and writes it
// to a timestamped PNG in the working directory, returning the path. It is a
// no-op (empty path) when no world view is showing.
func (g *Game) screenshot() (string, error) {
	w := g.activeWorld()
	if w == nil {
		return "", nil
	}
	g.renderer.SetHUD(false)
	fb := g.renderer.Frame(w)
	g.renderer.SetHUD(true) // play always shows the HUD; restore it

	cfg := g.renderer.Config()
	img := &image.RGBA{
		Pix:    append([]byte(nil), fb...),
		Stride: cfg.Width * 4,
		Rect:   image.Rect(0, 0, cfg.Width, cfg.Height),
	}
	name := fmt.Sprintf("pandemonium-%d.png", time.Now().UnixMilli())
	f, err := os.Create(name)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	if err := png.Encode(f, img); err != nil {
		return "", err
	}
	return name, nil
}
