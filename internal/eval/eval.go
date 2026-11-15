package eval

import (
	"fmt"
	"strings"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/sim/bot"
	"github.com/danielriddell21/pandemonium/internal/world"
)

const tickRate = 60

type Config struct {
	Runs          int
	Levels        int
	Width, Height int
	Skill         sim.Skill
	LevelSeconds  int
	FirstSeed     int64
}

func DefaultConfig() Config {
	return Config{Runs: 24, Levels: 5, Width: 48, Height: 32, Skill: sim.SkillNormal, LevelSeconds: 240, FirstSeed: 1}
}

type LevelResult struct {
	Seed    int64
	Depth   int
	Boss    bool
	Cleared bool
	Deaths  int
	Alarms  int
	Seconds float64
	Par     float64
	Damage  float64
}

type Report struct {
	Levels   []LevelResult
	Unsolved []int64
}

type tally struct {
	deaths int
}

func (t *tally) Observe(o sim.Observation) {
	if o.Kind == sim.ObsDeath {
		t.deaths++
	}
}

func Run(cfg Config) Report {
	var rep Report
	for run := range cfg.Runs {
		// Runs are spaced so their level seeds never overlap.
		base := cfg.FirstSeed + int64(run*cfg.Levels)
		loadout, carried := sim.Loadout{}, false
		for depth := range cfg.Levels {
			seed := base + int64(depth)
			l := generate(cfg, seed, world.IsArenaLevel(depth+1))
			if !solvable(cfg, l) {
				rep.Unsolved = append(rep.Unsolved, seed)
			}
			opts := []sim.Option{sim.WithSkill(cfg.Skill), sim.WithDepth(depth)}
			if carried {
				opts = append(opts, sim.WithLoadout(loadout))
			}
			res, g := play(cfg, l, opts)
			res.Seed, res.Depth = seed, depth
			rep.Levels = append(rep.Levels, res)
			if !res.Cleared {
				break // a stalled run has no next floor to carry into
			}
			loadout, carried = g.Loadout(), true
		}
	}
	return rep
}

func generate(cfg Config, seed int64, arena bool) *world.Level {
	l, err := world.Generate(world.Config{Width: cfg.Width, Height: cfg.Height, Seed: seed, Arena: arena})
	if err != nil {
		panic(fmt.Errorf("eval: seed %d: %w", seed, err))
	}
	return l
}

func solvable(cfg Config, l *world.Level) bool {
	t := &tally{}
	g := sim.New(l, sim.WithObserver(t))
	g.Entities = nil // no opposition: this asks only whether the level can be walked
	p := bot.Roamer(false)
	for range cfg.LevelSeconds * tickRate {
		g.Tick(p.Input(g), 1.0/tickRate)
		if g.LevelComplete() {
			return t.deaths == 0
		}
	}
	return false
}

func play(cfg Config, l *world.Level, opts []sim.Option) (LevelResult, *sim.Game) {
	t := &tally{}
	g := sim.New(l, append(opts, sim.WithObserver(t))...)
	p := bot.Roamer(true)
	res := LevelResult{Boss: l.Boss}
	for range cfg.LevelSeconds * tickRate {
		before := g.Player.Health + g.Player.Armor
		deaths := t.deaths
		g.Tick(p.Input(g), 1.0/tickRate)
		if lost := before - g.Player.Health - g.Player.Armor; lost > 0 && t.deaths == deaths {
			res.Damage += lost
		}
		if g.LevelComplete() {
			res.Cleared = true
			break
		}
	}
	st := g.LevelStats()
	res.Deaths, res.Alarms, res.Seconds, res.Par = t.deaths, st.Alarms, st.Elapsed, st.Par
	return res, g
}

type Metrics struct {
	Solvable   float64
	ClearRate  float64
	DeathRate  float64
	Damage     float64
	Alarms     float64
	Pace       float64
	BossClear  float64
	BossTime   float64
	Levels     int
	BossLevels int
}

func (r Report) Metrics() Metrics {
	var m Metrics
	var cleared, bossCleared int
	for _, l := range r.Levels {
		m.Levels++
		m.DeathRate += float64(l.Deaths)
		m.Damage += l.Damage
		if l.Boss {
			m.BossLevels++
			if l.Cleared {
				bossCleared++
				m.BossTime += l.Seconds
			}
			continue
		}
		m.Alarms += float64(l.Alarms)
		if l.Cleared {
			cleared++
			m.Pace += l.Seconds / l.Par
		}
	}
	if m.Levels == 0 {
		return m
	}
	n := float64(m.Levels)
	m.Solvable = 1 - float64(len(r.Unsolved))/n
	m.ClearRate = float64(cleared+bossCleared) / n
	m.DeathRate /= n
	m.Damage /= n
	if maze := m.Levels - m.BossLevels; maze > 0 {
		m.Alarms /= float64(maze)
	}
	if cleared > 0 {
		m.Pace /= float64(cleared)
	}
	if m.BossLevels > 0 {
		m.BossClear = float64(bossCleared) / float64(m.BossLevels)
	}
	if bossCleared > 0 {
		m.BossTime /= float64(bossCleared)
	}
	return m
}

type Band struct {
	Name     string
	Min, Max float64
	Value    func(Metrics) float64
}

func (b Band) Holds(m Metrics) bool {
	v := b.Value(m)
	return v >= b.Min && v <= b.Max
}

func Bands() []Band {
	return []Band{
		{"solvable (no demons)", 1, 1, func(m Metrics) float64 { return m.Solvable }},
		{"clear rate", 0.9, 1, func(m Metrics) float64 { return m.ClearRate }},
		{"deaths per level", 0.02, 0.4, func(m Metrics) float64 { return m.DeathRate }},
		{"damage taken per level", 20, 110, func(m Metrics) float64 { return m.Damage }},
		{"alarms per maze level", 0.15, 1.2, func(m Metrics) float64 { return m.Alarms }},
		{"clear time / par", 0.6, 1.6, func(m Metrics) float64 { return m.Pace }},
		{"boss floors cleared", 0.75, 1, func(m Metrics) float64 { return m.BossClear }},
		{"boss fight seconds", 20, 150, func(m Metrics) float64 { return m.BossTime }},
	}
}

func (m Metrics) Failures() []string {
	var out []string
	for _, b := range Bands() {
		if !b.Holds(m) {
			out = append(out, fmt.Sprintf("%s = %.2f, want %.2f..%.2f", b.Name, b.Value(m), b.Min, b.Max))
		}
	}
	return out
}

func (m Metrics) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d levels (%d boss floors)\n", m.Levels, m.BossLevels)
	for _, b := range Bands() {
		mark := "ok"
		if !b.Holds(m) {
			mark = "FAIL"
		}
		fmt.Fprintf(&sb, "  %-24s %7.2f   band %6.2f..%-6.2f %s\n", b.Name, b.Value(m), b.Min, b.Max, mark)
	}
	return sb.String()
}
