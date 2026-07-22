package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

type Records struct {
	Runs         int   `json:"runs"`
	DeepestLevel int   `json:"deepest_level"`
	Deaths       int   `json:"deaths"`
	Kills        int   `json:"kills"`
	PlaytimeMS   int64 `json:"playtime_ms"`
}

const reachCarryback = 4

func (r Records) Reach() int {
	if reach := r.DeepestLevel - reachCarryback; reach > 0 {
		return reach
	}
	return 0
}

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

func recordsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(dir, "pandemonium", "records.json"), nil
}

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
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal records: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write records: %w", err)
	}
	return nil
}

type RecordKeeper struct {
	base Records
	cur  Records
	path string
}

var _ telemetry.Subscriber = (*RecordKeeper)(nil)

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

func (k *RecordKeeper) Current() Records { return k.cur }

func (k *RecordKeeper) OnEvent(telemetry.PlayerEvent) {}

func (k *RecordKeeper) OnPathSummary(telemetry.PathSummary) {}

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
