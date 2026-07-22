package gui

import (
	"strings"
	"testing"
)

// The menu mechanics (move, select, adjust) now live in crucible/menu and are
// covered by its own tests; these check pandemonium's value formatting and
// the settings-row value column.

func TestValueFormatters(t *testing.T) {
	if got := percent(0.5); got != "50%" {
		t.Errorf("percent(0.5) = %q, want 50%%", got)
	}
	if got := onOff(true); got != "ON" {
		t.Errorf("onOff(true) = %q, want ON", got)
	}
	if got := onOff(false); got != "OFF" {
		t.Errorf("onOff(false) = %q, want OFF", got)
	}
	if got := times(1.5); got != "1.5x" {
		t.Errorf("times(1.5) = %q, want 1.5x", got)
	}
}

func TestAdjustRowShowsValueColumn(t *testing.T) {
	it := adjustRow("SOUND", func() string { return onOff(true) }, func(int) {})
	got := it.Label()
	if !strings.HasPrefix(got, "SOUND") || !strings.HasSuffix(got, "ON") {
		t.Errorf("adjustRow label = %q, want SOUND...ON", got)
	}
	if it.Adjust == nil {
		t.Error("adjustRow should carry an Adjust")
	}
}
