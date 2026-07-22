package status

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

func TestReporterImplementsSubscriber(_ *testing.T) {
	var _ telemetry.Subscriber = New(hud.New(), NewTableSource())
}

func TestReporterDiagnosticFromStart(t *testing.T) {
	// From the very first level the only thing posted is a debug-only telemetry
	// readout — never a player-facing notice (the narrator stays quiet early).
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnEvent(telemetry.PlayerEvent{Type: "exit", LevelIndex: 0})
	_, ch, ok := o.Active()
	if !ok || ch != hud.Diagnostic {
		t.Errorf("early band should post a Diagnostic readout, got ch=%v ok=%v", ch, ok)
	}
}

func TestReporterReachLiftsEarlyBand(t *testing.T) {
	// A run that carries reach from earlier progress speaks on an early level
	// that would otherwise be silent — the commentary resumes where it left off.
	o := hud.New()
	r := New(o, NewTableSource(), WithReach(7))
	r.OnEvent(telemetry.PlayerEvent{Type: "exit", LevelIndex: 0})
	if _, ch, ok := o.Active(); !ok || ch != hud.Notice {
		t.Error("with reach carried, an early exit should post a Notice")
	}
}

func TestReporterReachZeroIsPristine(t *testing.T) {
	// Zero reach keeps the early band free of any player-facing notice; only the
	// debug-only telemetry readout is posted.
	o := hud.New()
	r := New(o, NewTableSource(), WithReach(0))
	r.OnEvent(telemetry.PlayerEvent{Type: "exit", LevelIndex: 0})
	if _, ch, ok := o.Active(); ok && ch == hud.Notice {
		t.Error("zero reach must not surface a player notice in the early band")
	}
}

func TestReporterDiagnosticInMidBand(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnEvent(telemetry.PlayerEvent{Type: "exit", LevelIndex: 3})
	msg, ch, ok := o.Active()
	if !ok || ch != hud.Diagnostic {
		t.Errorf("mid band exit: got %q ch=%v ok=%v; want a Diagnostic message", msg, ch, ok)
	}
}

func TestReporterNoticeInLateBand(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnEvent(telemetry.PlayerEvent{Type: "exit", LevelIndex: 6})
	msg, ch, ok := o.Active()
	if !ok || ch != hud.Notice {
		t.Errorf("late band exit: got %q ch=%v ok=%v; want a Notice message", msg, ch, ok)
	}
}

func TestReporterWrongDoorNotice(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnEvent(telemetry.PlayerEvent{
		Type:       "door",
		LevelIndex: 7,
		Marker:     &telemetry.MarkerInfo{Kind: "door", WrongDoor: true},
	})
	if _, ch, ok := o.Active(); !ok || ch != hud.Notice {
		t.Error("wrong door in late band should post a Notice")
	}
}

func TestReporterIgnoresMovement(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnEvent(telemetry.PlayerEvent{Type: "move", LevelIndex: 9})
	if _, _, ok := o.Active(); ok {
		t.Error("movement events should not post messages")
	}
}

func TestReporterReactsToRepeatedWrongDoors(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnRunProfile(telemetry.RunProfile{LevelsCleared: 6, TotalWrongDoors: 4})
	r.OnEvent(telemetry.PlayerEvent{
		Type:       "door",
		LevelIndex: 6,
		Marker:     &telemetry.MarkerInfo{Kind: "door", WrongDoor: true},
	})
	msg, ch, ok := o.Active()
	if !ok || ch != hud.Notice || msg != "You keep opening the wrong ones." {
		t.Errorf("got %q ch=%v ok=%v; want the repeated-wrong-door notice", msg, ch, ok)
	}
}

func TestReporterReactsToRushing(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnRunProfile(telemetry.RunProfile{LevelsCleared: 6, ExploreScore: 0.2}) // low → rushing
	r.OnEvent(telemetry.PlayerEvent{Type: "exit", LevelIndex: 6})
	if msg, _, ok := o.Active(); !ok || msg != "Straight to the exit. Predictable." {
		t.Errorf("got %q ok=%v; want the rushing exit notice", msg, ok)
	}
}

func TestReporterLateBandIsMorePointed(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	// No profile context; late band (>=9) should still shift the wording.
	r.OnEvent(telemetry.PlayerEvent{
		Type:       "door",
		LevelIndex: 10,
		Marker:     &telemetry.MarkerInfo{Kind: "door", WrongDoor: true},
	})
	if msg, _, ok := o.Active(); !ok || msg != "You knew. You opened it anyway." {
		t.Errorf("got %q ok=%v; want the late-band wrong-door notice", msg, ok)
	}
}

func TestForkOptimalDetection(t *testing.T) {
	r := New(hud.New(), NewTableSource())
	cue, ok := r.cueFor(telemetry.PlayerEvent{
		Type:       "marker",
		LevelIndex: 4,
		Marker:     &telemetry.MarkerInfo{Kind: "junction", TakenX: 3, TakenY: 5, OptimalX: 3, OptimalY: 5},
	})
	if !ok || cue.Kind != CueFork || !cue.OptimalChoice {
		t.Errorf("expected optimal fork cue, got %+v ok=%v", cue, ok)
	}
}
