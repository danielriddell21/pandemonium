package gui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "settings.json")
	in := Settings{SFXVolume: 0.4, AmbientVolume: 0.7, Sensitivity: 1.6, FOV: 1.3, Crosshair: true}
	if err := in.saveTo(path); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := loadSettingsFrom(path)
	if got != in {
		t.Errorf("round trip mismatch: got %+v want %+v", got, in)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	got := loadSettingsFrom(filepath.Join(t.TempDir(), "nope.json"))
	if got != DefaultSettings() {
		t.Errorf("missing file: got %+v want defaults", got)
	}
}

func TestLoadCorruptFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := loadSettingsFrom(path); got != DefaultSettings() {
		t.Errorf("corrupt file: got %+v want defaults", got)
	}
}

func TestLoadClampsOutOfRangeValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wild.json")
	if err := os.WriteFile(path, []byte(`{"sfx_volume":9,"ambient_volume":-3,"mouse_sensitivity":99,"fov":0.01}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got := loadSettingsFrom(path)
	if got.SFXVolume != 1 || got.AmbientVolume != 0 {
		t.Errorf("volumes not clamped: %+v", got)
	}
	if got.Sensitivity > 3 || got.FOV < 0.6 {
		t.Errorf("sensitivity/fov not clamped: %+v", got)
	}
}
