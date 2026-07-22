package status

import (
	"strings"
	"testing"

	"github.com/danielriddell21/crucible/hud"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

func telemetryEvent(kind sim.ObservationKind, level int) telemetry.PlayerEvent {
	return telemetry.PlayerEvent{Kind: kind, LevelIndex: level}
}

func TestBandThresholds(t *testing.T) {
	cases := []struct {
		level int
		want  int
	}{
		{0, 0}, {1, 0}, {2, 1}, {4, 1}, {5, 2}, {8, 2}, {9, 3}, {14, 3}, {15, 4}, {30, 4},
	}
	for _, c := range cases {
		if got := band(c.level); got != c.want {
			t.Errorf("band(%d) = %d, want %d", c.level, got, c.want)
		}
	}
}

func TestScriptedLineByBand(t *testing.T) {
	cases := []struct {
		name    string
		cue     Cue
		wantOK  bool
		wantCh  hud.Channel
		wantHas string // substring the text must contain (optional)
	}{
		{"band0 diag exit", Cue{Kind: CueExit, Level: 0}, true, hud.Diagnostic, "telemetry"},
		{"band0 diag kill", Cue{Kind: CueKill, Level: 0}, true, hud.Diagnostic, "telemetry"},
		// band 1 — diagnostic readouts for every non-whispered cue
		{"band1 diag exit", Cue{Kind: CueExit, Level: 3}, true, hud.Diagnostic, "telemetry"},
		{"band1 diag exit rushing", Cue{Kind: CueExit, Level: 3, Rushing: true}, true, hud.Diagnostic, "rush"},
		{"band1 diag wrong door", Cue{Kind: CueWrongDoor, Level: 3}, true, hud.Diagnostic, "dead-end"},
		{"band1 diag decoy", Cue{Kind: CueDecoy, Level: 3}, true, hud.Diagnostic, "decoy"},
		{"band1 diag fork optimal", Cue{Kind: CueFork, Level: 3, OptimalChoice: true}, true, hud.Diagnostic, "optimal"},
		{"band1 diag fork suboptimal", Cue{Kind: CueFork, Level: 3}, true, hud.Diagnostic, "suboptimal"},
		{"band1 diag kill", Cue{Kind: CueKill, Level: 3}, true, hud.Diagnostic, "telemetry"},
		{"band1 diag item", Cue{Kind: CueItem, Level: 3}, true, hud.Diagnostic, "telemetry"},
		// band 1 — whispers for the charged moments
		{"band1 whisper death", Cue{Kind: CueDeath, Level: 3}, true, hud.Notice, "Again"},
		{"band1 whisper secret", Cue{Kind: CueSecret, Level: 3}, true, hud.Notice, "tucked away"},
		// band 2 — player-facing notices
		{"band2 exit", Cue{Kind: CueExit, Level: 6}, true, hud.Notice, "exit"},
		{"band2 exit rushing", Cue{Kind: CueExit, Level: 6, Rushing: true}, true, hud.Notice, "Predictable"},
		{"band2 wrong door", Cue{Kind: CueWrongDoor, Level: 6}, true, hud.Notice, "Nothing"},
		{"band2 repeated wrong doors", Cue{Kind: CueWrongDoor, Level: 6, WrongDoorsTotal: 4}, true, hud.Notice, "keep"},
		{"band2 decoy", Cue{Kind: CueDecoy, Level: 6}, true, hud.Notice, "onward"},
		{"band2 kill", Cue{Kind: CueKill, Level: 6}, true, hud.Notice, "Efficient"},
		{"band2 item", Cue{Kind: CueItem, Level: 6}, true, hud.Notice, "take what"},
		{"band2 secret", Cue{Kind: CueSecret, Level: 6}, true, hud.Notice, "Noted"},
		{"band2 death", Cue{Kind: CueDeath, Level: 6}, true, hud.Notice, "Back"},
		// band 3 — pointed
		{"band3 exit", Cue{Kind: CueExit, Level: 10}, true, hud.Notice, "keep finding"},
		{"band3 wrong door", Cue{Kind: CueWrongDoor, Level: 10}, true, hud.Notice, "You knew"},
		{"band3 decoy", Cue{Kind: CueDecoy, Level: 10}, true, hud.Notice, "not quite"},
		{"band3 kill", Cue{Kind: CueKill, Level: 10}, true, hud.Notice, "firing"},
		{"band3 item", Cue{Kind: CueItem, Level: 10}, true, hud.Notice, "changes anything"},
		{"band3 secret", Cue{Kind: CueSecret, Level: 10}, true, hud.Notice, "seam"},
		{"band3 death", Cue{Kind: CueDeath, Level: 10}, true, hud.Notice, "door either"},
		{"forks stay silent to the player", Cue{Kind: CueFork, Level: 10}, false, 0, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			line, ok := scriptedLine(c.cue)
			if ok != c.wantOK {
				t.Fatalf("ok = %v, want %v (line %q)", ok, c.wantOK, line.Text)
			}
			if !ok {
				return
			}
			if line.Channel != c.wantCh {
				t.Errorf("channel = %v, want %v", line.Channel, c.wantCh)
			}
			if c.wantHas != "" && !strings.Contains(line.Text, c.wantHas) {
				t.Errorf("text %q does not contain %q", line.Text, c.wantHas)
			}
			if line.Frames != messageFrames {
				t.Errorf("frames = %d, want %d", line.Frames, messageFrames)
			}
		})
	}
}

func TestDeathLineEscalatesWithCount(t *testing.T) {
	line, ok := scriptedLine(Cue{Kind: CueDeath, Level: 10, Deaths: 5})
	if !ok || !strings.Contains(line.Text, "so many times") {
		t.Errorf("repeated deaths should escalate, got %q ok=%v", line.Text, ok)
	}
}

func TestReporterOnPathSummaryIsInert(t *testing.T) {
	o := hud.New()
	r := New(o, NewTableSource())
	r.OnPathSummary(telemetry.PathSummary{})
	if _, _, ok := o.Active(); ok {
		t.Error("OnPathSummary should not post a message")
	}
}

func TestKillAndItemThrottledToFirstPerLevel(t *testing.T) {
	r := New(hud.New(), NewTableSource())
	first := func(kind sim.ObservationKind) bool {
		_, ok := r.cueFor(telemetryEvent(kind, 6))
		return ok
	}
	if !first(sim.ObsKill) {
		t.Error("first kill of a level should cue")
	}
	if first(sim.ObsKill) {
		t.Error("second kill of the same level should be throttled")
	}
	if !first(sim.ObsItem) {
		t.Error("first item of a level should cue")
	}
	if first(sim.ObsItem) {
		t.Error("second item of the same level should be throttled")
	}
	// A new level resets the throttle.
	if _, ok := r.cueFor(telemetryEvent(sim.ObsKill, 7)); !ok {
		t.Error("a new level should allow a kill cue again")
	}
}
