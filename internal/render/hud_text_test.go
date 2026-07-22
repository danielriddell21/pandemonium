package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func changedPixels(cfg Config, before, after []byte) (band, below int) {
	for y := range cfg.Height {
		for x := range cfg.Width {
			i := (y*cfg.Width + x) * 4
			if before[i] != after[i] || before[i+1] != after[i+1] || before[i+2] != after[i+2] {
				if y < 16 {
					band++
				} else {
					below++
				}
			}
		}
	}
	return band, below
}

func TestNoticeMessageDrawn(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)

	o := hud.New()
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	r := NewRenderer(cfg, WithOverlay(o))

	before := append([]byte(nil), r.Frame(g)...)
	o.Post("THE WAY IS OPEN", 60, hud.Notice)
	after := r.Frame(g)

	band, below := changedPixels(cfg, before, after)
	if band == 0 {
		t.Error("expected notice text pixels in the top HUD band")
	}
	if below != 0 {
		t.Errorf("notice text should stay in the HUD band, changed %d pixels below", below)
	}
}

func TestDiagnosticHiddenWhenDisabled(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)

	o := hud.New()
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	r := NewRenderer(cfg, WithOverlay(o), WithDiagnostics(false))

	before := append([]byte(nil), r.Frame(g)...)
	o.Post("telemetry: backtracking", 60, hud.Diagnostic)
	after := r.Frame(g)

	if band, below := changedPixels(cfg, before, after); band != 0 || below != 0 {
		t.Errorf("diagnostic message must not render when disabled (changed %d/%d)", band, below)
	}
}

func TestDiagnosticShownWhenEnabled(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 3})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)

	o := hud.New()
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	r := NewRenderer(cfg, WithOverlay(o), WithDiagnostics(true))

	before := append([]byte(nil), r.Frame(g)...)
	o.Post("telemetry: backtracking", 60, hud.Diagnostic)
	after := r.Frame(g)

	if band, _ := changedPixels(cfg, before, after); band == 0 {
		t.Error("diagnostic message should render when enabled")
	}
}

func TestNoOverlayDeterministic(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 2})
	if err != nil {
		t.Fatal(err)
	}
	g := sim.New(l)
	cfg := Config{Width: 160, Height: 100, FOV: 1.152}

	r := NewRenderer(cfg)
	a := append([]byte(nil), r.Frame(g)...)
	b := r.Frame(g)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("frame differs without overlay at byte %d", i)
		}
	}
}
