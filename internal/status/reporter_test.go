package status

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

func TestReporterImplementsSubscriber(_ *testing.T) {
	var _ telemetry.Subscriber = New(hud.New(), NewTableSource())
}

func TestReporterSilentInEarlyBand(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnEvent(telemetry.PlayerEvent{Type: "exit", LevelIndex: 0})
	if _, _, ok := o.Active(); ok {
		t.Error("expected silence in the early band (pure game)")
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

func TestForkOptimalDetection(t *testing.T) {
	cue, ok := cueFor(telemetry.PlayerEvent{
		Type:       "marker",
		LevelIndex: 4,
		Marker:     &telemetry.MarkerInfo{Kind: "junction", TakenX: 3, TakenY: 5, OptimalX: 3, OptimalY: 5},
	})
	if !ok || cue.Kind != CueFork || !cue.OptimalChoice {
		t.Errorf("expected optimal fork cue, got %+v ok=%v", cue, ok)
	}
}
