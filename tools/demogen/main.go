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
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/sim"
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
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
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
	level, err := world.Generate(world.Config{Width: c.mapW, Height: c.mapH, Seed: c.seed})
	if err != nil {
		return 0, err
	}
	g := sim.New(level)
	if c.setup != nil {
		c.setup(g)
	}
	r := render.NewRenderer(c.rcfg)
	r.SetAutomap(c.automap)

	anim := &gif.GIF{LoopCount: 0}
	var raw [][]byte // collected RGBA frames, for video encoding
	emit := func(rgba []byte) {
		if c.video {
			cp := make([]byte, len(rgba))
			copy(cp, rgba)
			raw = append(raw, cp)
			return
		}
		anim.Image = append(anim.Image, toPaletted(rgba, c.rcfg, pal))
		anim.Delay = append(anim.Delay, c.delayCs)
	}

	seen := map[world.Coord]bool{}
	nextSeed := c.seed
	const dt = 1.0 / 12.0
	subSteps := max(c.subSteps, 1)

	for i := range c.frames {
		reached := false
		for range subSteps {
			g.Tick(c.input(i, g), dt)
			if g.LevelComplete() {
				reached = true
				break
			}
		}
		if reached {
			// Hold on the level-complete tally before moving on, when asked.
			if c.tally {
				stats := g.LevelStats()
				for range 18 {
					emit(r.Intermission(stats))
				}
			}
			nextSeed++
			nl, err := world.Generate(world.Config{Width: c.mapW, Height: c.mapH, Seed: nextSeed})
			if err != nil {
				return 0, err
			}
			g = sim.New(nl)
		}
		seen[g.PlayerCell()] = true
		emit(r.Frame(g))
	}

	if c.video {
		return len(seen), encodeMP4(path, raw, c.rcfg, c.delayCs)
	}

	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	if err := gif.EncodeAll(f, anim); err != nil {
		return 0, err
	}
	return len(seen), nil
}

// encodeMP4 pipes raw RGBA frames to ffmpeg and writes an H.264 MP4. The frame
// rate is derived from the per-frame GIF delay so video and GIF clips play at the
// same speed.
func encodeMP4(path string, frames [][]byte, rcfg render.Config, delayCs int) error {
	if len(frames) == 0 {
		return fmt.Errorf("no frames to encode")
	}
	fps := 100.0 / float64(delayCs)
	cmd := exec.Command("ffmpeg",
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
	return cmd.Run()
}

func defaultClips() []clip {
	cfg := render.Config{Width: 256, Height: 160, FOV: 1.152}
	return []clip{
		{name: "hero", seed: 16, mapW: 40, mapH: 26, rcfg: cfg, frames: 360, delayCs: 7, input: pathFollow(true), tally: true, video: true},
		{name: "exploration", seed: 12, mapW: 32, mapH: 24, rcfg: cfg, frames: 84, delayCs: 7, input: pathFollow(false)},
		combatClip(cfg),
		{name: "automap", seed: 7, mapW: 40, mapH: 26, rcfg: cfg, frames: 120, delayCs: 7, input: pathFollow(false), automap: true},
		terrainClip(cfg),
	}
}

// terrainClip finds a level with a long staircase and walks slowly up it, showing
// the stepped floors, the raised landing and the varied ceilings.
func terrainClip(cfg render.Config) clip {
	const w, h = 32, 24
	type spot struct {
		seed int64
		x, y int
		dir  world.Coord
	}
	var best spot
	bestRise := 0.0
	dirs := []world.Coord{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}}
	for seed := int64(0); seed < 80; seed++ {
		l, err := world.Generate(world.Config{Width: w, Height: h, Seed: seed})
		if err != nil {
			continue
		}
		for y := 1; y < h-1; y++ {
			for x := 1; x < w-1; x++ {
				if !l.At(x, y).Walkable() {
					continue
				}
				for _, d := range dirs {
					rise, ok := 0.0, true
					for k := 1; k <= 4; k++ {
						nx, ny := x+d.X*k, y+d.Y*k
						if !l.At(nx, ny).Walkable() {
							ok = false
							break
						}
						rise = l.Floor(nx, ny) - l.Floor(x, y)
					}
					if ok && rise > bestRise {
						bestRise = rise
						best = spot{seed: seed, x: x, y: y, dir: d}
					}
				}
			}
		}
	}
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

// --- inputs ---------------------------------------------------------------

// pathFollow steers along the shortest route to the exit, detouring to grab the
// keycard that unlocks it, opening doors in its way, and recovering if a death
// respawns it. When fight is set it also breaks off to shoot demons in reach;
// when not, it simply walks the level (a calm exploration clip). Walking a whole
// level also shows it regenerating when the exit is reached.
func pathFollow(fight bool) func(int, *sim.Game) sim.Input {
	var (
		level   *world.Level
		route   []world.Coord
		idx     int
		lastPos sim.Vec2
		stuck   int
	)
	plan := func(g *sim.Game) {
		level = g.World.Level
		route = planRoute(g)
		idx = 0
	}
	return func(_ int, g *sim.Game) sim.Input {
		if g.World.Level != level {
			plan(g)
		}
		if len(route) == 0 {
			return sim.Input{Forward: 1}
		}

		// Break off to fight a demon that has come within reach, so the run
		// actually clears some of the level rather than strolling past them.
		if fight {
			if e, ok := engageDemon(g); ok {
				return engage(g, e)
			}
		}

		// Advance past waypoints we've reached. Use the occupied cell, not exact
		// centring, since a body on a raised step can't always sit dead-centre.
		for idx < len(route)-1 && (g.PlayerCell() == route[idx] || dist(g.Player.Pos, centre(route[idx])) < 0.5) {
			idx++
		}
		target := centre(route[idx])
		if dist(g.Player.Pos, target) > 3 { // teleported (e.g. a death respawn)
			plan(g)
			target = centre(route[idx])
		}

		// If we've barely moved for a while, we're snagged on a corner or step —
		// skip the troublesome waypoint and strafe to slip free.
		var nudge float64
		if dist(g.Player.Pos, lastPos) < 0.02 {
			if stuck++; stuck > 10 {
				if idx < len(route)-1 {
					idx++
					target = centre(route[idx])
				}
				nudge = 1
				stuck = 0
			}
		} else {
			stuck = 0
		}
		lastPos = g.Player.Pos

		// At the end of the route, turn to the exit switch and throw it.
		if idx == len(route)-1 {
			if sw, ok := exitSwitchCell(g.World.Level); ok {
				target = centre(sw)
			}
		}

		// Steer toward the waypoint by sliding on both axes, not just walking the
		// way we face — this rounds corners and stepped corridors without snagging.
		in := steerToward(g, target)
		in.Strafe += nudge
		openDoorAhead(g, &in)
		if fight {
			attackIfThreatened(g, &in)
		}
		return in
	}
}

// engageDemon returns the nearest living demon within a short fighting range and
// clear line of sight, so the path follower can deal with it before moving on.
func engageDemon(g *sim.Game) (sim.Entity, bool) {
	const engageRange = 5.0
	best := math.Inf(1)
	var found sim.Entity
	ok := false
	for _, e := range g.Entities {
		if !e.Alive {
			continue
		}
		d := dist(e.Pos, g.Player.Pos)
		if d > engageRange || !losClear(g.World, g.Player.Pos, e.Pos) {
			continue
		}
		if d < best {
			best, found, ok = d, e, true
		}
	}
	return found, ok
}

// engage turns to face a demon, closes to firing distance and shoots it.
func engage(g *sim.Game, e sim.Entity) sim.Input {
	diff := angleDiff(math.Atan2(e.Pos.Y-g.Player.Pos.Y, e.Pos.X-g.Player.Pos.X), g.Player.Angle)
	var in sim.Input
	if math.Abs(diff) > 0.04 {
		in.TurnDelta = clampF(diff, -0.16, 0.16)
	}
	if math.Abs(diff) < 0.4 && dist(e.Pos, g.Player.Pos) > 2.2 {
		in.Forward = 1
	}
	attackIfThreatened(g, &in)
	return in
}

// planRoute builds the cells to walk: detour through the exit's keycard first (so
// the locked door will open when reached), then on to the exit. Doors are treated
// as passable for routing; the bot opens them as it arrives.
func planRoute(g *sim.Game) []world.Coord {
	l := g.World.Level
	start := g.PlayerCell()
	var route []world.Coord
	if key, ok := firstKeyCell(l); ok {
		if seg := bfsOpen(l, start, key); len(seg) > 0 {
			route = seg
			start = key
		}
	}
	if seg := bfsOpen(l, start, l.Exit); len(seg) > 0 {
		if len(route) > 0 {
			seg = seg[1:] // drop the shared cell where the segments meet
		}
		route = append(route, seg...)
	}
	return route
}

// firstKeyCell returns the location of the first keycard item on the level.
func firstKeyCell(l *world.Level) (world.Coord, bool) {
	for _, it := range l.Items {
		if it.Kind.IsKey() {
			return it.At, true
		}
	}
	return world.Coord{}, false
}

// openDoorAhead issues Interact when an unopened door or a switch sits directly
// in front of the player, mirroring how the simulation decides what the player
// can act on.
func openDoorAhead(g *sim.Game, in *sim.Input) {
	dir := g.Player.Dir()
	tx := int(math.Floor(g.Player.Pos.X + dir.X*0.9))
	ty := int(math.Floor(g.Player.Pos.Y + dir.Y*0.9))
	if _, ok := g.World.Level.SwitchAt(tx, ty); ok {
		in.Interact = true
	}
	if g.World.IsDoor(tx, ty) && !g.World.Opened(tx, ty) {
		in.Interact = true
	}
}

// steerToward turns to face a target and moves toward it on both the forward and
// strafe axes, so the bot slides around corners and up stepped corridors instead
// of stalling when it can't walk in a dead-straight line.
func steerToward(g *sim.Game, target sim.Vec2) sim.Input {
	toX, toY := target.X-g.Player.Pos.X, target.Y-g.Player.Pos.Y
	desired := math.Atan2(toY, toX)
	var in sim.Input
	in.TurnDelta = clampF(angleDiff(desired, g.Player.Angle), -0.2, 0.2)
	dir := g.Player.Dir()
	fwd := toX*dir.X + toY*dir.Y    // component along facing
	str := toX*(-dir.Y) + toY*dir.X // component along the strafe axis
	in.Forward = clampF(fwd, -1, 1)
	in.Strafe = clampF(str, -1, 1)
	return in
}

// exitSwitchCell returns the cell of the level's exit switch, if it has one.
func exitSwitchCell(l *world.Level) (world.Coord, bool) {
	for c, s := range l.Switches {
		if s.Action == world.SwitchExit {
			return c, true
		}
	}
	return world.Coord{}, false
}

// combatClip looks for a level that opens on a demon in view and is well stocked
// with others, then turns the bot loose to fight its way through them.
func combatClip(cfg render.Config) clip {
	const w, h = 32, 24
	for seed := int64(0); seed < 400; seed++ {
		l, err := world.Generate(world.Config{Width: w, Height: h, Seed: seed})
		if err != nil {
			continue
		}
		g := sim.New(l)
		target, ok := visibleDemon(g)
		if !ok || len(g.Entities) < 5 {
			continue
		}
		angle := math.Atan2(target.Y-g.Player.Pos.Y, target.X-g.Player.Pos.X)
		return clip{
			name: "combat", seed: seed, mapW: w, mapH: h, rcfg: cfg,
			frames: 110, delayCs: 7,
			setup: func(g *sim.Game) { g.Player.Angle = angle },
			input: hunt(),
		}
	}
	return clip{name: "combat", seed: 5, mapW: w, mapH: h, rcfg: cfg, frames: 110, delayCs: 7, input: hunt()}
}

// hunt works through the level's demons one after another: it shoots whatever it
// can see, and navigates toward the nearest survivor when none is in sight, so the
// clip is a sustained fight rather than a single kill.
func hunt() func(int, *sim.Game) sim.Input {
	var (
		lvl    *world.Level
		path   []world.Coord
		idx    int
		toCell world.Coord
	)
	return func(_ int, g *sim.Game) sim.Input {
		e, ok := nearestDemon(g)
		if !ok {
			return sim.Input{} // all clear — hold while the last kill settles
		}
		// Fight anything in sight; otherwise close the gap toward it.
		if losClear(g.World, g.Player.Pos, e.Pos) && dist(e.Pos, g.Player.Pos) <= 6 {
			return engage(g, e)
		}

		ec := world.Coord{X: int(e.Pos.X), Y: int(e.Pos.Y)}
		if g.World.Level != lvl || ec != toCell || idx >= len(path) {
			lvl, toCell = g.World.Level, ec
			path = bfsOpen(lvl, g.PlayerCell(), ec)
			idx = 0
		}
		if len(path) == 0 {
			return sim.Input{Forward: 1}
		}
		target := centre(path[idx])
		for dist(g.Player.Pos, target) < 0.35 && idx < len(path)-1 {
			idx++
			target = centre(path[idx])
		}
		diff := angleDiff(math.Atan2(target.Y-g.Player.Pos.Y, target.X-g.Player.Pos.X), g.Player.Angle)
		var in sim.Input
		if math.Abs(diff) > 0.05 {
			in.TurnDelta = clampF(diff, -0.18, 0.18)
		}
		if math.Abs(diff) < 0.7 {
			in.Forward = 1
		}
		openDoorAhead(g, &in)
		return in
	}
}

// attackIfThreatened sets Attack when a living demon is just ahead in range.
func attackIfThreatened(g *sim.Game, in *sim.Input) {
	dir := g.Player.Dir()
	for _, e := range g.Entities {
		if !e.Alive {
			continue
		}
		dx, dy := e.Pos.X-g.Player.Pos.X, e.Pos.Y-g.Player.Pos.Y
		d := math.Hypot(dx, dy)
		if d == 0 || d > 2.9 {
			continue
		}
		if (dx/d)*dir.X+(dy/d)*dir.Y < 0.9 {
			continue
		}
		if losClear(g.World, g.Player.Pos, e.Pos) {
			in.Attack = true
			return
		}
	}
}

// --- setups ---------------------------------------------------------------

// --- helpers --------------------------------------------------------------

// bfsOpen finds the shortest cell path from src to dst over a level, treating
// every door as passable (walls and the world edge are the only obstacles). It is
// what the demo bot uses to route through locked doors it intends to open.
func bfsOpen(l *world.Level, src, dst world.Coord) []world.Coord {
	passable := func(c world.Coord) bool {
		return l.InBounds(c.X, c.Y) && l.At(c.X, c.Y).Walkable()
	}
	// climbable mirrors the simulation's step rule: a body can step up at most
	// world.MaxStep and drop any distance, so the bot only routes where it can
	// actually walk.
	climbable := func(from, to world.Coord) bool {
		return l.Floor(to.X, to.Y)-l.Floor(from.X, from.Y) <= world.MaxStep+1e-9
	}
	if !passable(src) || !passable(dst) {
		return nil
	}
	prev := map[world.Coord]world.Coord{src: src}
	queue := []world.Coord{src}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == dst {
			break
		}
		for _, n := range neighbours(c) {
			if !passable(n) || !climbable(c, n) {
				continue
			}
			if _, seen := prev[n]; seen {
				continue
			}
			prev[n] = c
			queue = append(queue, n)
		}
	}
	if _, ok := prev[dst]; !ok {
		return nil
	}
	var rev []world.Coord
	for c := dst; c != src; c = prev[c] {
		rev = append(rev, c)
	}
	rev = append(rev, src)
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

func visibleDemon(g *sim.Game) (sim.Vec2, bool) {
	best := math.Inf(1)
	var pos sim.Vec2
	ok := false
	for _, e := range g.Entities {
		if !e.Alive {
			continue
		}
		d := dist(g.Player.Pos, e.Pos)
		if d < 2.5 || d > 7 || !losClear(g.World, g.Player.Pos, e.Pos) {
			continue
		}
		if d < best {
			best, pos, ok = d, e.Pos, true
		}
	}
	return pos, ok
}

func nearestDemon(g *sim.Game) (sim.Entity, bool) {
	best := math.Inf(1)
	var found sim.Entity
	ok := false
	for _, e := range g.Entities {
		if !e.Alive {
			continue
		}
		if d := dist(e.Pos, g.Player.Pos); d < best {
			best, found, ok = d, e, true
		}
	}
	return found, ok
}

func losClear(w *sim.World, a, b sim.Vec2) bool {
	steps := int(dist(a, b)/0.1) + 1
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := a.X + (b.X-a.X)*t
		y := a.Y + (b.Y-a.Y)*t
		if w.Solid(int(math.Floor(x)), int(math.Floor(y))) {
			return false
		}
	}
	return true
}

func neighbours(c world.Coord) [4]world.Coord {
	return [4]world.Coord{
		{X: c.X + 1, Y: c.Y}, {X: c.X - 1, Y: c.Y},
		{X: c.X, Y: c.Y + 1}, {X: c.X, Y: c.Y - 1},
	}
}

func centre(c world.Coord) sim.Vec2 { return sim.Vec2{X: float64(c.X) + 0.5, Y: float64(c.Y) + 0.5} }
func dist(a, b sim.Vec2) float64    { return math.Hypot(a.X-b.X, a.Y-b.Y) }

func angleDiff(target, current float64) float64 {
	d := math.Mod(target-current+math.Pi, 2*math.Pi)
	if d < 0 {
		d += 2 * math.Pi
	}
	return d - math.Pi
}

func clampF(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// toPaletted quantises an RGBA frame buffer to the demo palette.
func toPaletted(fb []byte, cfg render.Config, pal color.Palette) *image.Paletted {
	rect := image.Rect(0, 0, cfg.Width, cfg.Height)
	rgba := &image.RGBA{Pix: fb, Stride: cfg.Width * 4, Rect: rect}
	p := image.NewPaletted(rect, pal)
	draw.Draw(p, rect, rgba, image.Point{}, draw.Src)
	return p
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
	pal := color.Palette{color.RGBA{A: 255}}
	const steps = 18
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
