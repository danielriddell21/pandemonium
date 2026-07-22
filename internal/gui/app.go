package gui

import (
	"fmt"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/sim/bot"
)

type state int

const (
	stateTitle state = iota

	statePlaying

	stateIntermission

	statePaused

	stateSettings

	stateAttract
)

const attractDelay = 600

type NextFunc func() *sim.Game

type Game struct {
	sim      *sim.Game
	renderer *render.Renderer
	overlay  *hud.Overlay
	start    NextFunc
	next     NextFunc
	setSkill func(skill int)

	audio     *Audio
	wasFiring bool

	state      state
	resumeTo   state
	menuFrom   state
	tally      sim.LevelStats
	showMap    bool
	depth      int
	haveMouse  bool
	lastMouseX int
	quit       bool

	settings     Settings
	records      *RecordKeeper
	titleMenu    *menuModel
	pauseMenu    *menuModel
	settingsMenu *menuModel

	attractGen   func() *sim.Game
	attract      *sim.Game
	attractPilot *bot.Pilot
	idle         int
}

var _ ebiten.Game = (*Game)(nil)

type Option func(*Game)

func WithOverlay(o *hud.Overlay) Option {
	return func(g *Game) { g.overlay = o }
}

func WithAudio(a *Audio) Option {
	return func(g *Game) { g.audio = a }
}

func WithSettings(s Settings) Option {
	return func(g *Game) { g.settings = s.clamped() }
}

func WithAttract(gen func() *sim.Game) Option {
	return func(g *Game) { g.attractGen = gen }
}

func WithRecords(k *RecordKeeper) Option {
	return func(g *Game) { g.records = k }
}

func WithDifficulty(set func(skill int)) Option {
	return func(g *Game) { g.setSkill = set }
}

func (g *Game) titleSubtitle() string {
	if g.records == nil {
		return ""
	}
	return g.records.Current().Summary()
}

func New(start, next NextFunc, renderer *render.Renderer, opts ...Option) *Game {
	game := &Game{start: start, next: next, renderer: renderer, state: stateTitle, settings: DefaultSettings()}
	for _, opt := range opts {
		opt(game)
	}
	game.buildMenus()
	game.applySettings()
	return game
}

func (g *Game) applySettings() {
	sfx, ambient := g.settings.SFXVolume, g.settings.AmbientVolume
	if !g.settings.Sound {
		sfx, ambient = 0, 0
	}
	g.audio.SetVolumes(sfx, ambient)
	g.renderer.SetFOV(g.settings.FOV)
	g.renderer.SetCrosshair(g.settings.Crosshair)
	g.renderer.SetDiagnostics(g.settings.Debug)
	ebiten.SetFullscreen(g.settings.Fullscreen)
	if g.setSkill != nil {
		g.setSkill(g.settings.Difficulty)
	}
}

func (g *Game) applyCursor() {
	if g.state == statePlaying {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}
}

func (g *Game) Update() error {
	if g.quit {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		if path, err := g.screenshot(); err != nil {
			fmt.Fprintln(os.Stderr, "screenshot:", err)
		} else if path != "" {
			fmt.Println("saved screenshot:", path)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		g.settings.Fullscreen = !g.settings.Fullscreen
		_ = g.settings.Save()
		ebiten.SetFullscreen(g.settings.Fullscreen)
	}
	g.applyCursor()
	switch g.state {
	case stateTitle:
		g.updateTitle()
	case stateAttract:
		g.updateAttract()
	case statePaused:
		g.updatePaused()
	case stateSettings:
		g.updateSettings()
	case stateIntermission:
		g.updateIntermission()
	case statePlaying:
		g.updatePlaying()
	}
	return nil
}

type nav struct {
	up, down, left, right, enter, back bool
}

func readNav() nav {
	return nav{
		up:    inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW),
		down:  inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS),
		left:  inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA),
		right: inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD),
		enter: inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace),
		back:  inpututil.IsKeyJustPressed(ebiten.KeyEscape),
	}
}

func (n nav) any() bool { return n.up || n.down || n.left || n.right || n.enter || n.back }

func (g *Game) menuNav() nav {
	n := readNav()
	if n.any() && g.audio != nil {
		g.audio.Menu()
	}
	return n
}

func (m *menuModel) drive(n nav) {
	switch {
	case n.up:
		m.move(-1)
	case n.down:
		m.move(1)
	case n.left:
		m.adjust(-1)
	case n.right:
		m.adjust(1)
	case n.enter:
		m.activate()
	}
}

func (g *Game) updateTitle() {
	n := g.menuNav()
	if n.any() {
		g.idle = 0
	} else if g.idle++; g.idle > attractDelay && g.attractGen != nil {
		g.startAttract()
		return
	}
	if n.back {
		g.quit = true
		return
	}
	g.titleMenu.drive(n)
}

func (g *Game) startAttract() {
	g.attract = g.attractGen()
	g.attractPilot = bot.Roamer(true)
	g.state = stateAttract
	g.idle = 0
}

func (g *Game) updateAttract() {
	if len(inpututil.AppendJustPressedKeys(nil)) > 0 ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.state = stateTitle
		return
	}
	dt := 1.0 / float64(ebiten.TPS())
	g.attract.Tick(g.attractPilot.Input(g.attract), dt)
	if g.attract.LevelComplete() {
		g.attract = g.attractGen()
		g.attractPilot = bot.Roamer(true)
	}
}

func (g *Game) updatePaused() {
	n := g.menuNav()
	if n.back {
		g.resume()
		return
	}
	g.pauseMenu.drive(n)
}

func (g *Game) resume() {
	g.state = g.resumeTo
	g.haveMouse = false // swallow the cursor jump accumulated while paused
}

func (g *Game) updateSettings() {
	n := g.menuNav()
	if n.back {
		g.closeSettings()
		return
	}
	g.settingsMenu.drive(n)
}

func (g *Game) closeSettings() {
	_ = g.settings.Save() // keep the in-memory values even if the disk write fails
	g.state = g.menuFrom
	g.haveMouse = false
}

func (g *Game) updateIntermission() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.resumeTo = stateIntermission
		g.state = statePaused
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if ng := g.next(); ng != nil {
			g.sim = ng
		}
		g.depth++
		g.renderer.SetGloom(1 - 0.04*float64(g.depth)) // the world darkens as the run deepens
		g.state = statePlaying
		g.haveMouse = false
	}
}

func (g *Game) updatePlaying() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.resumeTo = statePlaying
		g.state = statePaused
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.showMap = !g.showMap
		g.renderer.SetAutomap(g.showMap)
	}

	dt := 1.0 / float64(ebiten.TPS())
	if g.audio != nil {
		g.audio.SetListener(g.sim.Player.Pos) // attenuate the tick's sounds by distance
	}
	g.sim.Tick(g.readInput(), dt)

	// Play the weapon sound once per shot, on the muzzle-flash rising edge.
	firing := g.sim.MuzzleFlash()
	if firing && !g.wasFiring && g.audio != nil {
		g.audio.Fire()
	}
	g.wasFiring = firing

	if g.overlay != nil {
		g.overlay.Tick()
	}

	if g.sim.LevelComplete() {
		g.tally = g.sim.LevelStats()
		g.state = stateIntermission
	}
}

func (g *Game) buildMenus() {
	g.titleMenu = &menuModel{entries: []menuEntry{
		{label: "START", activate: func() {
			g.sim = g.start() // build the first level now, with the chosen difficulty
			g.state = statePlaying
			g.haveMouse = false
		}},
		{label: "SETTINGS", activate: func() {
			g.menuFrom = stateTitle
			g.state = stateSettings
		}},
		{label: "QUIT", activate: func() { g.quit = true }},
	}}

	g.pauseMenu = &menuModel{entries: []menuEntry{
		{label: "RESUME", activate: g.resume},
		{label: "SETTINGS", activate: func() {
			g.menuFrom = statePaused
			g.state = stateSettings
		}},
		{label: "QUIT", activate: func() { g.quit = true }},
	}}

	g.settingsMenu = &menuModel{entries: []menuEntry{
		{
			label: "DIFFICULTY",
			value: func() string { return sim.Skill(g.settings.Difficulty).String() },
			adjust: func(dir int) {
				g.settings.Difficulty = (g.settings.Difficulty + dir + skillCount) % skillCount
				g.applySettings()
			},
		},
		{
			label: "SOUND",
			value: func() string { return onOff(g.settings.Sound) },
			adjust: func(int) {
				g.settings.Sound = !g.settings.Sound
				g.applySettings()
			},
		},
		{
			label: "SFX VOLUME",
			value: func() string { return percent(g.settings.SFXVolume) },
			adjust: func(dir int) {
				g.settings.SFXVolume = clampRange(g.settings.SFXVolume+0.1*float64(dir), 0, 1)
				g.applySettings()
			},
		},
		{
			label: "AMBIENT VOLUME",
			value: func() string { return percent(g.settings.AmbientVolume) },
			adjust: func(dir int) {
				g.settings.AmbientVolume = clampRange(g.settings.AmbientVolume+0.1*float64(dir), 0, 1)
				g.applySettings()
			},
		},
		{
			label: "MOUSE SENSITIVITY",
			value: func() string { return times(g.settings.Sensitivity) },
			adjust: func(dir int) {
				g.settings.Sensitivity = clampRange(g.settings.Sensitivity+0.2*float64(dir), 0.2, 3)
			},
		},
		{
			label: "FIELD OF VIEW",
			value: func() string { return degrees(g.settings.FOV) },
			adjust: func(dir int) {
				step := 5 * math.Pi / 180
				g.settings.FOV = clampRange(g.settings.FOV+step*float64(dir), 0.6, 1.8)
				g.applySettings()
			},
		},
		{
			label: "CROSSHAIR",
			value: func() string { return onOff(g.settings.Crosshair) },
			adjust: func(int) {
				g.settings.Crosshair = !g.settings.Crosshair
				g.applySettings()
			},
		},
		{
			label: "DEBUG MESSAGES",
			value: func() string { return onOff(g.settings.Debug) },
			adjust: func(int) {
				g.settings.Debug = !g.settings.Debug
				g.applySettings()
			},
		},
		{
			label: "FULLSCREEN",
			value: func() string { return onOff(g.settings.Fullscreen) },
			adjust: func(int) {
				g.settings.Fullscreen = !g.settings.Fullscreen
				g.applySettings()
			},
		},
		{label: "BACK", activate: func() { g.closeSettings() }},
	}}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case stateTitle:
		screen.WritePixels(g.renderer.Menu("PANDEMONIUM", g.titleSubtitle(), g.titleMenu.items(), g.titleMenu.sel,
			"Up/Down select   Enter confirm"))
	case statePaused:
		screen.WritePixels(g.renderer.Menu("PAUSED", "", g.pauseMenu.items(), g.pauseMenu.sel,
			"Esc resumes"))
	case stateSettings:
		screen.WritePixels(g.renderer.Menu("SETTINGS", "", g.settingsMenu.items(), g.settingsMenu.sel,
			"Left/Right adjust   Esc back"))
	case stateAttract:
		screen.WritePixels(g.renderer.Frame(g.attract))
	case stateIntermission:
		screen.WritePixels(g.renderer.Intermission(g.tally))
	default:
		screen.WritePixels(g.renderer.Frame(g.sim))
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	cfg := g.renderer.Config()
	return cfg.Width, cfg.Height
}

const windowScale = 2

func (g *Game) Run() error {
	cfg := g.renderer.Config()
	ebiten.SetWindowSize(cfg.Width*windowScale, cfg.Height*windowScale)
	ebiten.SetWindowTitle("pandemonium")
	g.audio.StartAmbient()
	return g.handleRunError(ebiten.RunGame(g))
}
