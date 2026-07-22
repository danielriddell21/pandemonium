package gui

import (
	"path/filepath"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

func TestRecordsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "records.json")
	in := Records{Runs: 3, DeepestLevel: 12, Deaths: 7, Kills: 140, PlaytimeMS: 90000}
	if err := in.saveTo(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	if got := loadRecordsFrom(path); got != in {
		t.Errorf("round trip: got %+v want %+v", got, in)
	}
}

func TestLoadMissingRecordsIsZero(t *testing.T) {
	if got := loadRecordsFrom(filepath.Join(t.TempDir(), "none.json")); got != (Records{}) {
		t.Errorf("missing file: got %+v want zero", got)
	}
}

func TestReachCarriesBack(t *testing.T) {
	if got := (Records{DeepestLevel: 12}).Reach(); got != 12-reachCarryback {
		t.Errorf("reach = %d, want %d", got, 12-reachCarryback)
	}
	if got := (Records{DeepestLevel: 2}).Reach(); got != 0 {
		t.Errorf("shallow reach = %d, want 0", got)
	}
}

func TestSummaryEmptyForFirstRun(t *testing.T) {
	if got := (Records{}).Summary(); got != "" {
		t.Errorf("first-run summary = %q, want empty", got)
	}
	if got := (Records{Runs: 1, DeepestLevel: 0}).Summary(); got == "" {
		t.Error("a played run should produce a summary")
	}
}

func TestKeeperCountsRunAndFoldsProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.json")
	base := Records{Runs: 2, DeepestLevel: 5, Deaths: 3, Kills: 50}
	k := newRecordKeeperAt(base, path)

	// Creating the keeper counts the run immediately and persists it.
	if k.Current().Runs != 3 {
		t.Errorf("runs = %d, want 3", k.Current().Runs)
	}
	if loadRecordsFrom(path).Runs != 3 {
		t.Error("new run not persisted on construction")
	}

	// A profile deeper than the base lifts the deepest and adds to the totals.
	k.OnRunProfile(telemetry.RunProfile{LevelsCleared: 8, Deaths: 2, TotalKills: 30})
	cur := k.Current()
	if cur.DeepestLevel != 8 {
		t.Errorf("deepest = %d, want 8", cur.DeepestLevel)
	}
	if cur.Deaths != 5 || cur.Kills != 80 {
		t.Errorf("totals = deaths %d kills %d, want 5/80", cur.Deaths, cur.Kills)
	}

	// A shallower later profile must not lower the deepest reached.
	k.OnRunProfile(telemetry.RunProfile{LevelsCleared: 1, Deaths: 2, TotalKills: 30})
	if k.Current().DeepestLevel != 8 {
		t.Errorf("deepest dropped to %d, want 8", k.Current().DeepestLevel)
	}
}
