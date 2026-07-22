package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Settings struct {
	Difficulty    int     `json:"difficulty"`
	Sound         bool    `json:"sound"`
	SFXVolume     float64 `json:"sfx_volume"`
	AmbientVolume float64 `json:"ambient_volume"`
	Sensitivity   float64 `json:"mouse_sensitivity"`
	FOV           float64 `json:"fov"`
	Crosshair     bool    `json:"crosshair"`
	Debug         bool    `json:"debug"`
	Fullscreen    bool    `json:"fullscreen"`
}

const skillCount = 4

const defaultDifficulty = 1

const defaultFOV = 1.152

func DefaultSettings() Settings {
	return Settings{
		Difficulty:    defaultDifficulty,
		Sound:         true,
		SFXVolume:     1,
		AmbientVolume: 1,
		Sensitivity:   1,
		FOV:           defaultFOV,
		Crosshair:     false,
		Fullscreen:    true,
	}
}

func (s Settings) clamped() Settings {
	if s.Difficulty < 0 || s.Difficulty >= skillCount {
		s.Difficulty = defaultDifficulty
	}
	s.SFXVolume = clampRange(s.SFXVolume, 0, 1)
	s.AmbientVolume = clampRange(s.AmbientVolume, 0, 1)
	s.Sensitivity = clampRange(s.Sensitivity, 0.2, 3)
	s.FOV = clampRange(s.FOV, 0.6, 1.8)
	return s
}

func clampRange(v, lo, hi float64) float64 { return max(lo, min(hi, v)) }

func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(dir, "pandemonium", "settings.json"), nil
}

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

func (s Settings) Save() error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	return s.saveTo(path)
}

func (s Settings) saveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}
