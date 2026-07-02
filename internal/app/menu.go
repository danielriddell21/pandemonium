package app

import (
	"fmt"
	"math"

	"github.com/danielriddell21/pandemonium/internal/render"
)

// menuEntry is one row of a menu: a label, an optional live value, and the
// actions bound to it. All behaviour lives in closures so the model itself
// stays a plain, testable list.
type menuEntry struct {
	label    string
	value    func() string // current value shown beside the label, if any
	adjust   func(dir int) // left/right action (dir is -1 or +1), if any
	activate func()        // enter action, if any
}

// menuModel is a selectable list. It knows nothing about input devices or
// rendering; the app feeds it moves and reads it out as render items.
type menuModel struct {
	entries []menuEntry
	sel     int
}

// move shifts the selection by dy, wrapping at the ends.
func (m *menuModel) move(dy int) {
	n := len(m.entries)
	if n == 0 {
		return
	}
	m.sel = ((m.sel+dy)%n + n) % n
}

// adjust applies a left/right change to the selected entry, if it has one.
func (m *menuModel) adjust(dir int) {
	if e := m.entries[m.sel]; e.adjust != nil {
		e.adjust(dir)
	}
}

// activate fires the selected entry's action. Entries with an adjuster but no
// action treat enter as "adjust up", so toggles respond to both.
func (m *menuModel) activate() {
	e := m.entries[m.sel]
	switch {
	case e.activate != nil:
		e.activate()
	case e.adjust != nil:
		e.adjust(1)
	}
}

// items converts the model to the renderer's menu rows.
func (m *menuModel) items() []render.MenuItem {
	out := make([]render.MenuItem, len(m.entries))
	for i, e := range m.entries {
		it := render.MenuItem{Label: e.label}
		if e.value != nil {
			it.Value = e.value()
		}
		out[i] = it
	}
	return out
}

// percent formats a 0..1 value as a whole percentage.
func percent(v float64) string { return fmt.Sprintf("%d%%", int(math.Round(v*100))) }

// times formats a multiplier like "1.4x".
func times(v float64) string { return fmt.Sprintf("%.1fx", v) }

// degrees formats a radian angle as whole degrees.
func degrees(rad float64) string { return fmt.Sprintf("%d", int(math.Round(rad*180/math.Pi))) }

// onOff renders a toggle state.
func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}
