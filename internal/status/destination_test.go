package status

import (
	"strings"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

func TestArrivalFiresOnceThenVoiceGoesSparse(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())

	// First event in the deepest band (level >= 15): the terminal beat.
	r.OnEvent(telemetry.PlayerEvent{Kind: sim.ObsKill, LevelIndex: 16})
	msg, ch, ok := o.Active()
	if !ok || ch != hud.Notice || !strings.Contains(msg, "deep as it goes") {
		t.Fatalf("first deep-band event should fire the arrival line, got %q ok=%v", msg, ok)
	}
	o.Post("", 0, hud.Notice) // clear

	// Afterwards the chatty cues fall silent...
	r.OnEvent(telemetry.PlayerEvent{Kind: sim.ObsItem, LevelIndex: 16})
	if _, _, ok := o.Active(); ok {
		t.Error("after arrival, item pickups should no longer speak")
	}
	// ...but the big beats still land.
	r.OnEvent(telemetry.PlayerEvent{Kind: sim.ObsExit, LevelIndex: 16})
	if msg, _, ok := o.Active(); !ok || !strings.Contains(msg, "exactly like this one") {
		t.Errorf("after arrival, the exit should still draw a resigned line, got %q ok=%v", msg, ok)
	}
}

func TestDeepBandRegisterShift(t *testing.T) {
	// Below the deepest band the kill line is brisk; at it, it turns resigned.
	mid, _ := scriptedLine(Cue{Kind: CueKill, Level: 10})
	deep, _ := scriptedLine(Cue{Kind: CueKill, Level: 16})
	if mid.Text == deep.Text {
		t.Errorf("the deepest band should shift register, both said %q", mid.Text)
	}
	if !strings.Contains(deep.Text, "always more") {
		t.Errorf("deep-band kill line = %q, want the resigned variant", deep.Text)
	}
}
