package eval

import (
	"strings"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

func TestGate(t *testing.T) {
	rep := Run(DefaultConfig())
	m := rep.Metrics()
	t.Log("\n" + m.String())
	for _, seed := range rep.Unsolved {
		t.Errorf("seed %d cannot be walked start to finish with the demons removed", seed)
	}
	for _, f := range m.Failures() {
		t.Error(f)
	}
}

func TestRunIsDeterministic(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Runs, cfg.Levels = 2, 2
	a, b := Run(cfg), Run(cfg)
	if len(a.Levels) == 0 || len(a.Levels) != len(b.Levels) {
		t.Fatalf("runs differ in length: %d vs %d", len(a.Levels), len(b.Levels))
	}
	for i := range a.Levels {
		if a.Levels[i] != b.Levels[i] {
			t.Errorf("level %d differs between identical runs:\n%+v\n%+v", i, a.Levels[i], b.Levels[i])
		}
	}
}

func TestRunWalksTheCampaignInOrder(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Runs, cfg.Levels, cfg.FirstSeed = 2, 5, 100
	rep := Run(cfg)
	if len(rep.Levels) != 10 {
		t.Fatalf("want 10 levels, got %d", len(rep.Levels))
	}
	for i, l := range rep.Levels {
		if want := int64(100 + i); l.Seed != want {
			t.Errorf("level %d: seed %d, want %d", i, l.Seed, want)
		}
		if l.Depth != i%5 {
			t.Errorf("level %d: depth %d, want %d", i, l.Depth, i%5)
		}
		if l.Boss != (l.Depth == 4) {
			t.Errorf("level %d: boss=%v at depth %d", i, l.Boss, l.Depth)
		}
		if l.Par <= 0 || l.Seconds <= 0 {
			t.Errorf("level %d: par %v, seconds %v", i, l.Par, l.Seconds)
		}
	}
}

func TestStalledRunStopsThere(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Runs, cfg.Levels, cfg.LevelSeconds = 1, 3, 1 // a second is not enough to clear anything
	rep := Run(cfg)
	if len(rep.Levels) != 1 || rep.Levels[0].Cleared {
		t.Errorf("a run that stalls on its first floor should end there: %+v", rep.Levels)
	}
	if len(rep.Unsolved) != 1 {
		t.Errorf("a level not walked in time counts as unsolved: %v", rep.Unsolved)
	}
}

func TestHarderSkillHurtsMore(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Runs, cfg.Levels = 6, 2
	easy, hard := cfg, cfg
	easy.Skill, hard.Skill = sim.SkillEasy, sim.SkillNightmare
	e, h := Run(easy).Metrics(), Run(hard).Metrics()
	if h.Damage <= e.Damage {
		t.Errorf("nightmare (%.0f damage/level) should hurt more than easy (%.0f)", h.Damage, e.Damage)
	}
}

func TestMetrics(t *testing.T) {
	rep := Report{
		Levels: []LevelResult{
			{Cleared: true, Deaths: 1, Alarms: 2, Seconds: 30, Par: 60, Damage: 40},
			{Cleared: true, Alarms: 0, Seconds: 90, Par: 60, Damage: 20},
			{Cleared: false, Deaths: 3, Alarms: 1, Seconds: 240, Par: 60, Damage: 120},
			{Boss: true, Cleared: true, Seconds: 70, Par: 40, Damage: 100},
			{Boss: true, Cleared: false, Deaths: 4, Seconds: 240, Par: 40, Damage: 220},
		},
		Unsolved: []int64{7},
	}
	m := rep.Metrics()
	approx := func(name string, got, want float64) {
		t.Helper()
		if d := got - want; d > 1e-9 || d < -1e-9 {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}
	approx("solvable", m.Solvable, 0.8)
	approx("clear rate", m.ClearRate, 0.6)
	approx("death rate", m.DeathRate, 8.0/5)
	approx("damage", m.Damage, 100)
	approx("alarms", m.Alarms, 1) // over maze levels only
	approx("pace", m.Pace, 1)     // cleared maze levels only: (0.5 + 1.5) / 2
	approx("boss clear", m.BossClear, 0.5)
	approx("boss time", m.BossTime, 70)
	if m.Levels != 5 || m.BossLevels != 2 {
		t.Errorf("levels=%d boss=%d", m.Levels, m.BossLevels)
	}
}

func TestMetricsOfNothing(t *testing.T) {
	if m := (Report{}).Metrics(); m != (Metrics{}) {
		t.Errorf("an empty report should have zero metrics: %+v", m)
	}
	m := Report{Levels: []LevelResult{{Boss: true}}}.Metrics()
	if m.Pace != 0 || m.BossTime != 0 || m.Alarms != 0 {
		t.Errorf("no cleared levels should not divide by zero: %+v", m)
	}
}

func inBand() Metrics {
	return Metrics{Solvable: 1, ClearRate: 0.97, DeathRate: 0.1, Damage: 60, Alarms: 0.4, Pace: 0.9, BossClear: 1, BossTime: 60, Levels: 10, BossLevels: 2}
}

func TestBandsHaveAFloorAndACeiling(t *testing.T) {
	if f := inBand().Failures(); len(f) != 0 {
		t.Fatalf("a healthy game failed its bands: %v", f)
	}
	for _, b := range Bands() {
		if b.Min > b.Max {
			t.Errorf("band %q is inverted: %v..%v", b.Name, b.Min, b.Max)
		}
	}
	// Too hard and too easy both fail: the gate is a band, not a floor.
	tooHard := inBand()
	tooHard.DeathRate, tooHard.Damage, tooHard.ClearRate = 3, 400, 0.4
	if got := len(tooHard.Failures()); got != 3 {
		t.Errorf("a brutal game should fail three bands, failed %d: %v", got, tooHard.Failures())
	}
	tooEasy := inBand()
	tooEasy.DeathRate, tooEasy.Damage, tooEasy.Alarms, tooEasy.BossTime = 0, 2, 0, 5
	if got := len(tooEasy.Failures()); got != 4 {
		t.Errorf("a toothless game should fail four bands, failed %d: %v", got, tooEasy.Failures())
	}
	unsolved := inBand()
	unsolved.Solvable = 0.99
	if len(unsolved.Failures()) != 1 {
		t.Error("a single unwalkable level must fail the gate")
	}
}

func TestMetricsTable(t *testing.T) {
	s := inBand().String()
	for _, b := range Bands() {
		if !strings.Contains(s, b.Name) {
			t.Errorf("table is missing the %q row:\n%s", b.Name, s)
		}
	}
	if strings.Contains(s, "FAIL") {
		t.Errorf("a healthy table should not flag a failure:\n%s", s)
	}
	bad := inBand()
	bad.Damage = 999
	if !strings.Contains(bad.String(), "FAIL") {
		t.Error("an out-of-band row should be flagged")
	}
}

func TestGenerateRejectsNothingItIsAskedFor(t *testing.T) {
	cfg := DefaultConfig()
	if l := generate(cfg, 3, true); !l.Boss {
		t.Error("an arena request should produce a boss floor")
	}
	if l := generate(cfg, 3, false); l.Boss {
		t.Error("a maze request should not")
	}
}
