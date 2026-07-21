package app

import (
	"image/png"
	"os"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestScreenshotWritesPNGWhilePlaying(t *testing.T) {
	t.Chdir(t.TempDir())

	l, err := world.Generate(world.Config{Width: 24, Height: 18, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{
		sim:      sim.New(l),
		renderer: render.NewRenderer(render.Config{Width: 96, Height: 64, FOV: 1.152}),
		state:    statePlaying,
	}
	path, err := g.screenshot()
	if err != nil {
		t.Fatalf("screenshot: %v", err)
	}
	if path == "" {
		t.Fatal("expected a screenshot path while playing")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	if _, err := png.Decode(f); err != nil {
		t.Errorf("screenshot is not a valid PNG: %v", err)
	}
}

func TestScreenshotSkippedInMenus(t *testing.T) {
	g := &Game{
		renderer: render.NewRenderer(render.Config{Width: 96, Height: 64, FOV: 1.152}),
		state:    stateTitle,
	}
	if path, err := g.screenshot(); err != nil || path != "" {
		t.Errorf("menus should not produce a screenshot: path=%q err=%v", path, err)
	}
}
