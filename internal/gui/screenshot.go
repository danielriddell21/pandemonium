package gui

import (
	"fmt"
	"time"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

func (g *Game) activeWorld() *sim.Game {
	switch g.state {
	case statePlaying:
		return g.sim
	case stateAttract:
		return g.attract
	}
	return nil
}

func (g *Game) screenshot() (string, error) {
	w := g.activeWorld()
	if w == nil {
		return "", nil
	}
	g.renderer.SetHUD(false)
	fb := g.renderer.Frame(w)
	g.renderer.SetHUD(true) // play always shows the HUD; restore it

	cfg := g.renderer.Config()
	name := fmt.Sprintf("pandemonium-%d.png", time.Now().UnixMilli())
	if err := record.SavePNG(name, record.FromRGBA(fb, cfg.Width, cfg.Height)); err != nil {
		return "", fmt.Errorf("save screenshot: %w", err)
	}
	return name, nil
}
