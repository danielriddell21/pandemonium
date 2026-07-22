package gui

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func newTestGame(t *testing.T) *sim.Game {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 24, Height: 18, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	return sim.New(l)
}

func newApp(t *testing.T) *Game {
	t.Helper()
	r := render.NewRenderer(render.Config{Width: 160, Height: 100, FOV: 1.152})
	start := func() *sim.Game { return newTestGame(t) }
	next := func() *sim.Game { return newTestGame(t) }
	return New(start, next, r, WithAttract(func() *sim.Game { return newTestGame(t) }))
}

func TestStartsOnTitle(t *testing.T) {
	if got := newApp(t).state; got != stateTitle {
		t.Errorf("new game state = %d, want stateTitle", got)
	}
}

func TestTitleStartEntersPlaying(t *testing.T) {
	g := newApp(t)
	g.titleMenu.activate() // START is first
	if g.state != statePlaying {
		t.Errorf("after START, state = %d, want statePlaying", g.state)
	}
}

func TestTitleSettingsAndBack(t *testing.T) {
	g := newApp(t)
	g.titleMenu.move(1) // SETTINGS
	g.titleMenu.activate()
	if g.state != stateSettings {
		t.Fatalf("state = %d, want stateSettings", g.state)
	}
	g.closeSettings()
	if g.state != stateTitle {
		t.Errorf("after back, state = %d, want stateTitle", g.state)
	}
}

func TestPauseResumeReturnsToOrigin(t *testing.T) {
	g := newApp(t)
	g.resumeTo = statePlaying
	g.state = statePaused
	g.resume()
	if g.state != statePlaying {
		t.Errorf("resume state = %d, want statePlaying", g.state)
	}
}

func TestSettingsAdjustAppliesToRenderer(t *testing.T) {
	g := newApp(t)
	// Find the FOV entry and nudge it up; the renderer config must follow.
	before := g.renderer.Config().FOV
	for _, e := range g.settingsMenu.entries {
		if e.label == "FIELD OF VIEW" {
			e.adjust(1)
		}
	}
	after := g.renderer.Config().FOV
	if after <= before {
		t.Errorf("FOV did not increase: %.3f -> %.3f", before, after)
	}
	if g.settings.FOV != after {
		t.Errorf("settings FOV %.3f out of sync with renderer %.3f", g.settings.FOV, after)
	}
}

func TestCrosshairToggle(t *testing.T) {
	g := newApp(t)
	start := g.settings.Crosshair
	for _, e := range g.settingsMenu.entries {
		if e.label == "CROSSHAIR" {
			e.adjust(1)
		}
	}
	if g.settings.Crosshair == start {
		t.Error("crosshair setting did not toggle")
	}
}

func TestSoundAndDebugToggle(t *testing.T) {
	g := newApp(t)
	for _, name := range []string{"SOUND", "DEBUG MESSAGES"} {
		for _, e := range g.settingsMenu.entries {
			if e.label == name {
				e.adjust(1) // must flip and apply without panicking (audio is nil)
			}
		}
	}
	if g.settings.Sound { // default true, toggled once
		t.Error("SOUND did not toggle off")
	}
	if !g.settings.Debug { // default false, toggled once
		t.Error("DEBUG MESSAGES did not toggle on")
	}
}

func TestDifficultyCyclesAndReachesSink(t *testing.T) {
	got := -1
	r := render.NewRenderer(render.Config{Width: 160, Height: 100, FOV: 1.152})
	start := func() *sim.Game { return newTestGame(t) }
	g := New(start, start, r, WithDifficulty(func(d int) { got = d }))
	// applySettings on construction pushes the default (normal) to the sink.
	if got != defaultDifficulty {
		t.Fatalf("sink got %d on construction, want %d", got, defaultDifficulty)
	}
	for _, e := range g.settingsMenu.entries {
		if e.label == "DIFFICULTY" {
			e.adjust(1) // normal -> hard
		}
	}
	if g.settings.Difficulty != defaultDifficulty+1 {
		t.Errorf("difficulty = %d, want %d", g.settings.Difficulty, defaultDifficulty+1)
	}
	if got != g.settings.Difficulty {
		t.Errorf("sink got %d, want %d", got, g.settings.Difficulty)
	}
}

func TestDifficultyWrapsAtEnds(t *testing.T) {
	g := newApp(t)
	g.settings.Difficulty = 0 // easiest
	for _, e := range g.settingsMenu.entries {
		if e.label == "DIFFICULTY" {
			e.adjust(-1) // wrap down to the hardest
			if g.settings.Difficulty != skillCount-1 {
				t.Errorf("wrap down = %d, want %d", g.settings.Difficulty, skillCount-1)
			}
		}
	}
}

func TestStartBuildsTheFirstLevel(t *testing.T) {
	g := newApp(t)
	if g.sim != nil {
		t.Fatal("sim should be nil before START")
	}
	g.titleMenu.activate() // START
	if g.sim == nil || g.state != statePlaying {
		t.Error("START should build the first level and enter play")
	}
}

func TestStartAttractBuildsLevel(t *testing.T) {
	g := newApp(t)
	g.startAttract()
	if g.state != stateAttract {
		t.Fatalf("state = %d, want stateAttract", g.state)
	}
	if g.attract == nil || g.attractPilot == nil {
		t.Error("attract level or pilot not initialised")
	}
}

func TestSettingsValuesRender(t *testing.T) {
	g := newApp(t)
	items := g.settingsMenu.items()
	if len(items) == 0 {
		t.Fatal("no settings items")
	}
	for _, it := range items[:len(items)-1] { // last entry is BACK, no value
		if it.Value == "" {
			t.Errorf("settings row %q has no value", it.Label)
		}
	}
}
