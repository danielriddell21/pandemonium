// Command demogen renders the documentation's gameplay media: short,
// deterministic animated clips (GIF/MP4) and static feature stills (PNG). It
// drives the simulation and the headless renderer directly — the same render
// pipeline the game uses — so it needs no display.
//
// Regenerate every asset under docs/demos with:
//
//	just demos      // or: go run ./tools/demogen
package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/sim/bot"
	"github.com/danielriddell21/pandemonium/internal/world"
)

const outDir = "docs/demos"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demogen:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	for _, c := range defaultClips() {
		path := filepath.Join(outDir, c.name+c.ext())
		visited, err := c.record(path)
		if err != nil {
			return fmt.Errorf("%s: %w", c.name, err)
		}
		fmt.Printf("wrote %-26s %3d frames, %2d tiles visited\n", path, c.frames, visited)
	}
	return recordStills()
}

// clip is one recorded sequence: a level, a render size, and a per-frame input.
type clip struct {
	name    string
	seed    int64
	mapW    int
	mapH    int
	rcfg    render.Config
	frames  int
	delayCs int
	setup   func(*sim.Game)
	input   func(i int, g *sim.Game) sim.Input
	tally   bool // hold on the level-complete tally screen when the exit is reached
	// subSteps simulates this many ticks per recorded frame (default 1), so long
	// runs can advance further per frame and keep the GIF compact.
	subSteps int
	// video writes an MP4 (via ffmpeg) instead of a GIF, which stays small even
	// for a long clip.
	video bool
	// automap overlays the explored-level minimap for the whole clip.
	automap bool
	// arena builds the open set-piece room instead of the usual maze.
	arena bool
}

// ext is the output file extension for the clip's encoding.
func (c clip) ext() string {
	if c.video {
		return ".mp4"
	}
	return ".gif"
}

func (c clip) record(path string) (int, error) {
	pal := demoPalette()
	g, err := c.newLevel(c.seed)
	if err != nil {
		return 0, err
	}
	r := render.NewRenderer(c.rcfg)
	r.SetAutomap(c.automap)

	// GIF clips encode through crucible's recorder: the demo palette quantises
	// cleanly without dithering (WithFrameDiff keeps the nearest-colour draw),
	// and delta frames keep the files compact.
	rec := record.NewRecorder(0, 1, 0, record.WithPalette(pal), record.WithFrameDelay(c.delayCs), record.WithFrameDiff())
	var raw [][]byte // collected RGBA frames, for video encoding
	emit := func(rgba []byte) {
		if c.video {
			cp := make([]byte, len(rgba))
			copy(cp, rgba)
			raw = append(raw, cp)
			return
		}
		rec.Add(&image.RGBA{
			Pix:    rgba,
			Stride: c.rcfg.Width * 4,
			Rect:   image.Rect(0, 0, c.rcfg.Width, c.rcfg.Height),
		})
	}
	// Hold on the level-complete tally before moving on, when asked.
	emitTally := func(g *sim.Game) {
		if !c.tally {
			return
		}
		stats := g.LevelStats()
		for range 18 {
			emit(r.Intermission(stats))
		}
	}

	seen := map[world.Coord]bool{}
	nextSeed := c.seed
	const dt = 1.0 / 12.0
	subSteps := max(c.subSteps, 1)

	for i := range c.frames {
		if c.runSubSteps(g, i, subSteps, dt) {
			emitTally(g)
			nextSeed++
			ng, err := c.newLevel(nextSeed)
			if err != nil {
				return 0, err
			}
			g = ng
		}
		seen[g.PlayerCell()] = true
		emit(r.Frame(g))
	}

	if c.video {
		return len(seen), encodeMP4(path, raw, c.rcfg, c.delayCs)
	}
	if err := rec.Save(path); err != nil {
		return 0, fmt.Errorf("save gif: %w", err)
	}
	return len(seen), nil
}

// newLevel generates a fresh level for the clip and applies its staging (e.g. no
// demons), so that staging carries across level transitions.
func (c clip) newLevel(seed int64) (*sim.Game, error) {
	l, err := world.Generate(world.Config{Width: c.mapW, Height: c.mapH, Seed: seed, Arena: c.arena})
	if err != nil {
		return nil, fmt.Errorf("generate level: %w", err)
	}
	g := sim.New(l)
	if c.setup != nil {
		c.setup(g)
	}
	return g, nil
}

// runSubSteps advances the game by the clip's sub-steps for frame i, reporting
// whether the level was completed during them.
func (c clip) runSubSteps(g *sim.Game, i, subSteps int, dt float64) bool {
	for range subSteps {
		g.Tick(c.input(i, g), dt)
		if g.LevelComplete() {
			return true
		}
	}
	return false
}

// encodeMP4 pipes raw RGBA frames to ffmpeg and writes an H.264 MP4. The frame
// rate is derived from the per-frame GIF delay so video and GIF clips play at the
// same speed.
func encodeMP4(path string, frames [][]byte, rcfg render.Config, delayCs int) error {
	if len(frames) == 0 {
		return fmt.Errorf("no frames to encode")
	}
	fps := 100.0 / float64(delayCs)
	cmd := exec.CommandContext(context.Background(), "ffmpeg",
		"-y", "-loglevel", "error",
		"-f", "rawvideo", "-pixel_format", "rgba",
		"-video_size", fmt.Sprintf("%dx%d", rcfg.Width, rcfg.Height),
		"-framerate", fmt.Sprintf("%.4f", fps),
		"-i", "-",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-movflags", "+faststart",
		path,
	)
	var buf bytes.Buffer
	for _, fr := range frames {
		buf.Write(fr)
	}
	cmd.Stdin = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run ffmpeg: %w", err)
	}
	return nil
}

func defaultClips() []clip {
	cfg := render.Config{Width: 256, Height: 160, FOV: 1.152}
	return []clip{
		{name: "hero", seed: 16, mapW: 40, mapH: 26, rcfg: cfg, frames: 360, delayCs: 7, input: pathFollow(true), tally: true, video: true},
		{name: "exploration", seed: 12, mapW: 32, mapH: 24, rcfg: cfg, frames: 84, delayCs: 7, setup: noDemons, input: pathFollow(false)},
		{name: "arena", seed: 5, mapW: 40, mapH: 28, rcfg: cfg, frames: 110, delayCs: 7, input: hunt(), arena: true},
		{name: "automap", seed: 7, mapW: 40, mapH: 26, rcfg: cfg, frames: 120, delayCs: 7, input: pathFollow(false), automap: true},
		terrainClip(cfg),
	}
}

// terrainClip finds a level with a long staircase and walks slowly up it, showing
// the stepped floors, the raised landing and the varied ceilings.
func terrainClip(cfg render.Config) clip {
	const w, h = 32, 24
	best := findStaircase(w, h)
	angle := math.Atan2(float64(best.dir.Y), float64(best.dir.X))
	return clip{
		name: "terrain", seed: best.seed, mapW: w, mapH: h, rcfg: cfg,
		frames: 70, delayCs: 8,
		setup: func(g *sim.Game) {
			g.Entities = nil
			g.Player.Pos = sim.Vec2{X: float64(best.x) + 0.5, Y: float64(best.y) + 0.5}
			g.Player.Z = g.World.Level.Floor(best.x, best.y)
			g.Player.Angle = angle
		},
		input: func(i int, _ *sim.Game) sim.Input {
			if i < 12 {
				return sim.Input{} // hold so the staircase reads before the climb
			}
			return sim.Input{Forward: 0.6}
		},
	}
}

// stairSpot marks where a climbable staircase was found: the level seed, the
// start cell, and the direction the steps rise.
type stairSpot struct {
	seed int64
	x, y int
	dir  world.Coord
}

// findStaircase searches the first levels for the one with the longest rising run
// of steps, returning where it starts.
func findStaircase(w, h int) stairSpot {
	dirs := []world.Coord{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}}
	var best stairSpot
	bestRise := 0.0
	for seed := int64(0); seed < 80; seed++ {
		l, err := world.Generate(world.Config{Width: w, Height: h, Seed: seed})
		if err != nil {
			continue
		}
		if spot, rise := bestStairInLevel(l, seed, w, h, dirs); rise > bestRise {
			best, bestRise = spot, rise
		}
	}
	return best
}

// bestStairInLevel returns the start cell of the steepest four-step rise in the
// level and that rise.
func bestStairInLevel(l *world.Level, seed int64, w, h int, dirs []world.Coord) (stairSpot, float64) {
	var best stairSpot
	bestRise := 0.0
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			if !l.At(x, y).Walkable() {
				continue
			}
			for _, d := range dirs {
				if rise, ok := stairRise(l, x, y, d); ok && rise > bestRise {
					bestRise = rise
					best = stairSpot{seed: seed, x: x, y: y, dir: d}
				}
			}
		}
	}
	return best, bestRise
}

// stairRise measures the floor rise over four steps from (x, y) in direction d,
// reporting false if the run leaves walkable floor.
func stairRise(l *world.Level, x, y int, d world.Coord) (float64, bool) {
	rise := 0.0
	for k := 1; k <= 4; k++ {
		nx, ny := x+d.X*k, y+d.Y*k
		if !l.At(nx, ny).Walkable() {
			return 0, false
		}
		rise = l.Floor(nx, ny) - l.Floor(x, y)
	}
	return rise, true
}

// --- inputs ---------------------------------------------------------------

// pathFollow steers along the shortest route to the exit via the bot package's
// roamer, adapting it to the clip input signature.
func pathFollow(fight bool) func(int, *sim.Game) sim.Input {
	p := bot.Roamer(fight)
	return func(_ int, g *sim.Game) sim.Input { return p.Input(g) }
}

// noDemons clears a level's demons, so a clip shows the world itself rather than
// a fight (the exploration walk-through).
func noDemons(g *sim.Game) { g.Entities = nil }

// hunt seeks and destroys the level's demons via the bot package's hunter.
func hunt() func(int, *sim.Game) sim.Input {
	p := bot.Hunter()
	return func(_ int, g *sim.Game) sim.Input { return p.Input(g) }
}

// demoPalette builds brightness ramps for every colour the renderer uses so the
// textured frames quantise cleanly without dithering.
func demoPalette() color.Palette {
	bases := []color.RGBA{
		{R: 28, G: 26, B: 30, A: 255},    // ceiling
		{R: 44, G: 36, B: 30, A: 255},    // floor
		{R: 150, G: 110, B: 78, A: 255},  // wall
		{R: 58, G: 44, B: 34, A: 255},    // mortar
		{R: 120, G: 70, B: 60, A: 255},   // door
		{R: 168, G: 52, B: 44, A: 255},   // demon
		{R: 120, G: 40, B: 96, A: 255},   // demon
		{R: 240, G: 220, B: 60, A: 255},  // eyes
		{R: 222, G: 214, B: 188, A: 255}, // notice text
		{R: 60, G: 220, B: 50, A: 255},   // health (full)
		{R: 240, G: 40, B: 50, A: 255},   // health (low)
	}
	const steps = 18
	pal := make(color.Palette, 1, 1+steps*len(bases))
	pal[0] = color.RGBA{A: 255}
	for _, b := range bases {
		for s := range steps {
			f := 0.1 + 0.9*float64(s)/float64(steps-1)
			pal = append(pal, color.RGBA{
				R: uint8(float64(b.R) * f),
				G: uint8(float64(b.G) * f),
				B: uint8(float64(b.B) * f),
				A: 255,
			})
		}
	}
	return pal
}
