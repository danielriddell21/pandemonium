// Package app is the Ebiten application: it owns the window, the game loop and
// input, and draws each frame using the render package. It is the only place
// that imports Ebiten. Dependency direction: app -> render -> sim -> world.
package app

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

// state is the app's top-level mode.
type state int

const (
	// stateTitle is the opening menu; idling here starts the attract loop.
	stateTitle state = iota
	// statePlaying is a live level.
	statePlaying
	// stateIntermission is the tally screen between levels.
	stateIntermission
	// statePaused is the in-run menu over a suspended game.
	statePaused
	// stateSettings is the options list, reachable from title and pause.
	stateSettings
	// stateAttract is the bot-driven demo that plays when the title idles.
	stateAttract
)

// attractDelay is how many ticks the title sits unattended before the demo
// loop starts (~10 seconds at 60 TPS).
const attractDelay = 600

// NextFunc produces the simulation for the next level. The app calls it when the
// player reaches an exit, so level progression and per-level wiring stay outside
// the app.
type NextFunc func() *sim.Game

// Game implements ebiten.Game, driving the current simulation and rendering it.
type Game struct {
	sim      *sim.Game
	renderer *render.Renderer
	overlay  *hud.Overlay
	start    NextFunc        // builds the run's first level, lazily on START
	next     NextFunc        // advances to the following level
	setSkill func(skill int) // pushes the chosen difficulty to the level builder

	audio     *Audio
	wasFiring bool // muzzle-flash state last tick, for one shot-sound per shot

	state      state
	resumeTo   state          // what closing the pause menu returns to
	menuFrom   state          // what closing the settings menu returns to
	tally      sim.LevelStats // captured stats shown on the intermission screen
	showMap    bool           // automap overlay toggled with Tab
	depth      int            // levels advanced into the run, dimming the world
	haveMouse  bool
	lastMouseX int
	quit       bool

	settings     Settings
	records      *RecordKeeper
	titleMenu    *menuModel
	pauseMenu    *menuModel
	settingsMenu *menuModel

	attractGen   func() *sim.Game // builds a fresh level for the demo loop
	attract      *sim.Game
	attractPilot *bot.Pilot
	idle         int // ticks the title has sat without input
}

var _ ebiten.Game = (*Game)(nil)

// Option configures a Game.
type Option func(*Game)

// WithOverlay attaches a HUD overlay that the loop advances each frame.
func WithOverlay(o *hud.Overlay) Option {
	return func(g *Game) { g.overlay = o }
}

// WithAudio attaches the sound engine. A nil engine leaves the game silent.
func WithAudio(a *Audio) Option {
	return func(g *Game) { g.audio = a }
}

// WithSettings applies persisted player options.
func WithSettings(s Settings) Option {
	return func(g *Game) { g.settings = s.clamped() }
}

// WithAttract supplies a generator for the title screen's demo loop: a fresh,
// unobserved simulation the bot can play while the menu idles.
func WithAttract(gen func() *sim.Game) Option {
	return func(g *Game) { g.attractGen = gen }
}

// WithRecords attaches the cross-run history shown on the title screen.
func WithRecords(k *RecordKeeper) Option {
	return func(g *Game) { g.records = k }
}

// WithDifficulty supplies a sink that receives the chosen difficulty (0-based
// skill index) so the level builder can use it. It is called whenever the
// setting changes, before the next level is built.
func WithDifficulty(set func(skill int)) Option {
	return func(g *Game) { g.setSkill = set }
}

// titleSubtitle is the records readout under the title, if any history exists.
func (g *Game) titleSubtitle() string {
	if g.records == nil {
		return ""
	}
	return g.records.Current().Summary()
}

// New builds the application around the level builders. start builds the run's
// first level when the player chooses START (so a difficulty picked on the title
// takes effect); next advances to a fresh level when the player reaches an exit.
func New(start, next NextFunc, renderer *render.Renderer, opts ...Option) *Game {
	game := &Game{start: start, next: next, renderer: renderer, state: stateTitle, settings: DefaultSettings()}
	for _, opt := range opts {
		opt(game)
	}
	game.buildMenus()
	game.applySettings()
	return game
}

// applySettings pushes the current options into the engine pieces that consume
// them. It is cheap and safe to call after every change. Muting (sound off, or
// a zero volume) is applied live; the engine itself is only created at launch
// when sound is enabled.
func (g *Game) applySettings() {
	sfx, ambient := g.settings.SFXVolume, g.settings.AmbientVolume
	if !g.settings.Sound {
		sfx, ambient = 0, 0
	}
	g.audio.SetVolumes(sfx, ambient)
	g.renderer.SetFOV(g.settings.FOV)
	g.renderer.SetCrosshair(g.settings.Crosshair)
	g.renderer.SetDiagnostics(g.settings.Debug)
	if g.setSkill != nil {
		g.setSkill(g.settings.Difficulty)
	}
}

// Update advances whichever mode the app is in.
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

// nav is one tick's worth of menu input.
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

// any reports whether the tick carried any menu input at all.
func (n nav) any() bool { return n.up || n.down || n.left || n.right || n.enter || n.back }

// drive applies one tick of navigation to a menu.
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
	n := readNav()
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

// startAttract builds a fresh unobserved level and hands it to the bot.
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
	n := readNav()
	if n.back {
		g.resume()
		return
	}
	g.pauseMenu.drive(n)
}

// resume leaves the pause menu for whatever it interrupted.
func (g *Game) resume() {
	g.state = g.resumeTo
	g.haveMouse = false // swallow the cursor jump accumulated while paused
}

func (g *Game) updateSettings() {
	n := readNav()
	if n.back {
		g.closeSettings()
		return
	}
	g.settingsMenu.drive(n)
}

// closeSettings persists the options and returns to the invoking screen.
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

// buildMenus wires the title, pause and settings screens.
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
		{label: "BACK", activate: func() { g.closeSettings() }},
	}}
}

// Draw renders whichever screen the app is on.
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

// Layout fixes the internal resolution; Ebiten scales it to the window.
func (g *Game) Layout(_, _ int) (int, int) {
	cfg := g.renderer.Config()
	return cfg.Width, cfg.Height
}

// windowScale enlarges the internal resolution to a comfortable window size.
const windowScale = 2

// Run opens the window and runs the game loop until the player quits. It blocks.
func (g *Game) Run() error {
	cfg := g.renderer.Config()
	ebiten.SetWindowSize(cfg.Width*windowScale, cfg.Height*windowScale)
	ebiten.SetWindowTitle("pandemonium")
	g.audio.StartAmbient()
	return g.handleRunError(ebiten.RunGame(g))
}
