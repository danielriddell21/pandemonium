package gui

import (
	"fmt"

	"github.com/danielriddell21/crucible/store"
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
	path, err := store.Path("pandemonium", "settings.json")
	if err != nil {
		return "", fmt.Errorf("locate settings path: %w", err)
	}
	return path, nil
}

func LoadSettings() Settings {
	path, err := settingsPath()
	if err != nil {
		return DefaultSettings()
	}
	return loadSettingsFrom(path)
}

func loadSettingsFrom(path string) Settings {
	return store.Load(path, DefaultSettings()).clamped()
}

func (s Settings) Save() error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	return s.saveTo(path)
}

func (s Settings) saveTo(path string) error {
	if err := store.Save(path, s); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	return nil
}
