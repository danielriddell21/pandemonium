package phrasing

import (
	"context"
	"strings"
	"testing"

	"github.com/danielriddell21/narrata"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/status"
)

func TestEventForMapping(t *testing.T) {
	cases := []struct {
		cue  status.Cue
		want string
	}{
		{status.Cue{Kind: status.CueExit, Level: 6}, "level_complete"},
		{status.Cue{Kind: status.CueKill, Level: 6}, "enemy_killed"},
		{status.Cue{Kind: status.CueItem, Level: 6}, "item_taken"},
		{status.Cue{Kind: status.CueSecret, Level: 6}, "secret_found"},
		{status.Cue{Kind: status.CueDeath, Level: 6}, "player_died"},
		{status.Cue{Kind: status.CueWrongDoor, Level: 6}, "wrong_turn"},
		{status.Cue{Kind: status.CueDecoy, Level: 6}, "decoy"},
		{status.Cue{Kind: status.CueFork, Level: 6}, "fork"},
		{status.Cue{Kind: status.CueKind(200), Level: 6}, "event"},
		{status.Cue{Kind: status.CueExit, Level: 16, Arrival: true}, "arrival"},
	}
	for _, c := range cases {
		if got := eventFor(c.cue); got != c.want {
			t.Errorf("eventFor(%+v) = %q, want %q", c.cue, got, c.want)
		}
	}
}

func newSource(t *testing.T) *Source {
	t.Helper()
	s, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestNoticeIsGenerated(t *testing.T) {
	s := newSource(t)
	var line status.Line
	s.Request(status.Cue{Kind: status.CueExit, Level: 6, LevelsCleared: 6}, func(l status.Line) { line = l })
	if line.Channel != hud.Notice {
		t.Fatalf("notice should stay on the Notice channel, got %+v", line)
	}
	if !strings.Contains(line.Text, "7") {
		t.Errorf("generated line should mention the level; got %q", line.Text)
	}
}

func TestGenerateFallsBack(t *testing.T) {
	s := newSource(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if text, ok := s.generate(ctx, status.Cue{Kind: status.CueExit, Level: 6}); ok {
		t.Errorf("cancelled context should yield no line, got %q", text)
	}
}

func TestDiagnosticNotGenerated(t *testing.T) {
	s := newSource(t)
	var line status.Line
	s.Request(status.Cue{Kind: status.CueExit, Level: 3}, func(l status.Line) { line = l })
	if line.Channel != hud.Diagnostic {
		t.Fatalf("expected a Diagnostic line, got %+v", line)
	}
}

func TestSilentCueStaysSilent(t *testing.T) {
	s := newSource(t)
	emitted := false
	s.Request(status.Cue{Kind: status.CueFork, Level: 10}, func(status.Line) { emitted = true })
	if emitted {
		t.Error("a fork is silent to the player; nothing should be emitted")
	}
}

func TestEngineNarratesEvent(t *testing.T) {
	engine, err := narrata.New(narrata.Config{})
	if err != nil {
		t.Fatalf("narrata.New: %v", err)
	}
	defer func() { _ = engine.Close() }()

	res, err := engine.Generate(context.Background(), narrata.Request{
		Event:  eventFor(status.Cue{Kind: status.CueExit, Level: 6}),
		Data:   dataFor(status.Cue{Kind: status.CueExit, Level: 6}),
		Output: narrata.OutputText,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if res.Silent || res.Text == "" {
		t.Errorf("expected a non-empty line, got %q (silent=%v)", res.Text, res.Silent)
	}
}
