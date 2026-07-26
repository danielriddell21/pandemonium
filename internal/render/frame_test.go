package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func smokeGame(t *testing.T) *sim.Game {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	return sim.New(l)
}

// TestFrameIsFullyPaintedAndVaried checks the whole pipeline produces an opaque,
// non-uniform image (walls, floor, ceiling and HUD all contributing).
func TestFrameIsFullyPaintedAndVaried(t *testing.T) {
	r := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	fb := r.Frame(smokeGame(t))
	if len(fb) != 160*100*4 {
		t.Fatalf("frame length = %d", len(fb))
	}
	var rmin, rmax byte = 255, 0
	for i := 0; i < len(fb); i += 4 {
		if fb[i+3] != 255 {
			t.Fatalf("pixel %d not opaque", i/4)
		}
		if fb[i] < rmin {
			rmin = fb[i]
		}
		if fb[i] > rmax {
			rmax = fb[i]
		}
	}
	if rmax-rmin < 20 {
		t.Errorf("frame looks flat: red range %d..%d", rmin, rmax)
	}
}

// TestHUDTogglesStatusBar checks SetHUD changes the bottom status-bar region.
func TestHUDTogglesStatusBar(t *testing.T) {
	g := smokeGame(t)
	r := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	with := append([]byte(nil), r.Frame(g)...)
	r.SetHUD(false)
	without := r.Frame(g)

	// Sample a row inside the status-bar band; it must differ with the HUD off.
	y := 100 - StatusBarH/2
	differs := false
	for x := 0; x < 160; x++ {
		i := (y*160 + x) * 4
		if with[i] != without[i] || with[i+1] != without[i+1] || with[i+2] != without[i+2] {
			differs = true
			break
		}
	}
	if !differs {
		t.Error("turning the HUD off should change the status-bar region")
	}
}

// TestCrosshairPaintsCentre checks the crosshair toggle marks pixels near the
// centre of the view.
func TestCrosshairPaintsCentre(t *testing.T) {
	g := smokeGame(t)
	g.Entities = nil
	r := NewRenderer(Config{Width: 160, Height: 100, FOV: 1.152})
	off := append([]byte(nil), r.Frame(g)...)
	r.SetCrosshair(true)
	on := r.Frame(g)

	diff := 0
	for i := range off {
		if off[i] != on[i] {
			diff++
		}
	}
	if diff == 0 {
		t.Error("enabling the crosshair changed nothing")
	}
}

// TestIntermissionRenders checks the tally screen paints an opaque frame.
func TestIntermissionRenders(t *testing.T) {
	r := NewRenderer(DefaultConfig())
	fb := r.Intermission(sim.LevelStats{Kills: 3, KillsTotal: 5, Elapsed: 40, Par: 60})
	if len(fb) != DefaultConfig().Width*DefaultConfig().Height*4 {
		t.Fatalf("intermission length = %d", len(fb))
	}
	for i := 3; i < len(fb); i += 4 {
		if fb[i] != 255 {
			t.Fatalf("intermission pixel %d not opaque", i/4)
		}
	}
}

// TestMenuRenders checks a menu screen paints an opaque frame with the selection.
func TestMenuRenders(t *testing.T) {
	r := NewRenderer(DefaultConfig())
	items := []MenuItem{{Label: "START"}, {Label: "QUIT"}}
	fb := r.Menu("PANDEMONIUM", "1 run", items, 0, "Enter")
	for i := 3; i < len(fb); i += 4 {
		if fb[i] != 255 {
			t.Fatalf("menu pixel %d not opaque", i/4)
		}
	}
}
