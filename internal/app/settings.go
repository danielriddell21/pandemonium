package app

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
)

// Settings are the player-tunable options, persisted between sessions as JSON
// under the user's config directory.
type Settings struct {
	Sound         bool    `json:"sound"`             // whether to run with audio at all
	SFXVolume     float64 `json:"sfx_volume"`        // 0..1
	AmbientVolume float64 `json:"ambient_volume"`    // 0..1
	Sensitivity   float64 `json:"mouse_sensitivity"` // multiplier on the base turn rate
	FOV           float64 `json:"fov"`               // horizontal field of view, radians
	Crosshair     bool    `json:"crosshair"`
	Debug         bool    `json:"debug"` // show on-screen diagnostic (playtest) messages
}

// defaultFOV mirrors render.DefaultConfig — restated here so settings stay a
// plain data file with no render dependency.
const defaultFOV = 1.152

// DefaultSettings returns the out-of-the-box options.
func DefaultSettings() Settings {
	return Settings{
		Sound:         true,
		SFXVolume:     1,
		AmbientVolume: 1,
		Sensitivity:   1,
		FOV:           defaultFOV,
		Crosshair:     false,
	}
}

// clamped returns the settings with every field forced into its valid range, so
// a hand-edited or stale file cannot produce a broken game.
func (s Settings) clamped() Settings {
	s.SFXVolume = clampRange(s.SFXVolume, 0, 1)
	s.AmbientVolume = clampRange(s.AmbientVolume, 0, 1)
	s.Sensitivity = clampRange(s.Sensitivity, 0.2, 3)
	s.FOV = clampRange(s.FOV, 0.6, 1.8)
	return s
}

func clampRange(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// settingsPath returns the JSON file location under the user config directory.
func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pandemonium", "settings.json"), nil
}

// LoadSettings reads the persisted settings, falling back to the defaults when
// the file is missing or unreadable.
func LoadSettings() Settings {
	path, err := settingsPath()
	if err != nil {
		return DefaultSettings()
	}
	return loadSettingsFrom(path)
}

func loadSettingsFrom(path string) Settings {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultSettings()
	}
	s := DefaultSettings()
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings()
	}
	return s.clamped()
}

// Save persists the settings, creating the config directory if needed. Errors
// are returned but safe to ignore: the game keeps the in-memory values.
func (s Settings) Save() error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	return s.saveTo(path)
}

func (s Settings) saveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
