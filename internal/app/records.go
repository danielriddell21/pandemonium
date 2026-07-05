package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

// Records is the cross-run history shown on the title screen and carried forward
// into the commentary: how many runs have been played, the deepest level
// reached, and lifetime death and kill counts. It persists as JSON beside the
// settings file.
type Records struct {
	Runs         int   `json:"runs"`
	DeepestLevel int   `json:"deepest_level"`
	Deaths       int   `json:"deaths"`
	Kills        int   `json:"kills"`
	PlaytimeMS   int64 `json:"playtime_ms"`
}

// reachCarryback is how many levels back from the deepest reached the commentary
// resumes on a new run: enough to leave a short re-climb rather than pinning a
// veteran at the top from the first step.
const reachCarryback = 4

// Reach is the level floor a fresh run inherits from this history.
func (r Records) Reach() int {
	if reach := r.DeepestLevel - reachCarryback; reach > 0 {
		return reach
	}
	return 0
}

// Summary is a one-line title-screen readout, empty for a first-ever run.
func (r Records) Summary() string {
	if r.Runs <= 0 {
		return ""
	}
	runs := "runs"
	if r.Runs == 1 {
		runs = "run"
	}
	return fmt.Sprintf("%d %s   deepest: level %d", r.Runs, runs, r.DeepestLevel+1)
}

// recordsPath returns the JSON file location under the user config directory.
func recordsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pandemonium", "records.json"), nil
}

// LoadRecords reads the persisted history, returning a zero value when the file
// is missing or unreadable.
func LoadRecords() Records {
	path, err := recordsPath()
	if err != nil {
		return Records{}
	}
	return loadRecordsFrom(path)
}

func loadRecordsFrom(path string) Records {
	data, err := os.ReadFile(path)
	if err != nil {
		return Records{}
	}
	var r Records
	if err := json.Unmarshal(data, &r); err != nil {
		return Records{}
	}
	return r
}

func (r Records) saveTo(path string) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// RecordKeeper folds the live run profile into the persisted history. It is a
// telemetry subscriber, so it updates as the run progresses, and it counts the
// run from the moment it is created (so even a run quit on level one is logged).
type RecordKeeper struct {
	base Records // history before this run
	cur  Records // history including this run so far
	path string
}

// Compile-time check that the keeper consumes telemetry.
var _ telemetry.Subscriber = (*RecordKeeper)(nil)

// NewRecordKeeper starts tracking a run on top of base, immediately recording
// that a new run has begun.
func NewRecordKeeper(base Records) *RecordKeeper {
	path, _ := recordsPath()
	return newRecordKeeperAt(base, path)
}

func newRecordKeeperAt(base Records, path string) *RecordKeeper {
	k := &RecordKeeper{base: base, cur: base, path: path}
	k.cur.Runs = base.Runs + 1
	_ = k.cur.saveTo(k.path)
	return k
}

// Current returns the history including this run so far.
func (k *RecordKeeper) Current() Records { return k.cur }

// OnEvent is unused: the keeper works from the cumulative run profile.
func (k *RecordKeeper) OnEvent(telemetry.PlayerEvent) {}

// OnPathSummary is unused: the keeper works from the cumulative run profile.
func (k *RecordKeeper) OnPathSummary(telemetry.PathSummary) {}

// OnRunProfile folds the run's cumulative totals onto the carried-forward base
// and persists the result.
func (k *RecordKeeper) OnRunProfile(p telemetry.RunProfile) {
	k.cur.DeepestLevel = max(k.cur.DeepestLevel, p.LevelsCleared)
	k.cur.Deaths = k.base.Deaths + p.Deaths
	k.cur.Kills = k.base.Kills + p.TotalKills
	var playMS int64
	for _, path := range p.Paths {
		playMS += path.TimeSpentMS
	}
	k.cur.PlaytimeMS = k.base.PlaytimeMS + playMS
	_ = k.cur.saveTo(k.path)
}
