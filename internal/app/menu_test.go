package app

import "testing"

func TestMenuMoveWraps(t *testing.T) {
	m := &menuModel{entries: make([]menuEntry, 3)}
	m.move(-1)
	if m.sel != 2 {
		t.Errorf("move up from 0 should wrap to 2, got %d", m.sel)
	}
	m.move(1)
	if m.sel != 0 {
		t.Errorf("move down from 2 should wrap to 0, got %d", m.sel)
	}
}

func TestMenuActivateFiresSelected(t *testing.T) {
	fired := -1
	m := &menuModel{entries: []menuEntry{
		{label: "A", activate: func() { fired = 0 }},
		{label: "B", activate: func() { fired = 1 }},
	}}
	m.move(1)
	m.activate()
	if fired != 1 {
		t.Errorf("activated entry %d, want 1", fired)
	}
}

func TestMenuAdjustAndEnterToggle(t *testing.T) {
	on := false
	m := &menuModel{entries: []menuEntry{
		{label: "T", adjust: func(int) { on = !on }},
	}}
	m.adjust(1)
	if !on {
		t.Fatal("adjust did not toggle on")
	}
	// An entry with only an adjuster treats enter as adjust.
	m.activate()
	if on {
		t.Fatal("enter did not toggle off")
	}
}

func TestMenuItemsReflectValues(t *testing.T) {
	v := 50
	m := &menuModel{entries: []menuEntry{
		{label: "VOL", value: func() string { return percent(float64(v) / 100) }},
	}}
	if got := m.items()[0].Value; got != "50%" {
		t.Errorf("value = %q, want 50%%", got)
	}
}
