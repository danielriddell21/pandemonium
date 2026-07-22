package gui

import (
	"fmt"
	"math"

	"github.com/danielriddell21/pandemonium/internal/render"
)

type menuEntry struct {
	label    string
	value    func() string
	adjust   func(dir int)
	activate func()
}

type menuModel struct {
	entries []menuEntry
	sel     int
}

func (m *menuModel) move(dy int) {
	n := len(m.entries)
	if n == 0 {
		return
	}
	m.sel = ((m.sel+dy)%n + n) % n
}

func (m *menuModel) adjust(dir int) {
	if e := m.entries[m.sel]; e.adjust != nil {
		e.adjust(dir)
	}
}

func (m *menuModel) activate() {
	e := m.entries[m.sel]
	switch {
	case e.activate != nil:
		e.activate()
	case e.adjust != nil:
		e.adjust(1)
	}
}

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

func percent(v float64) string { return fmt.Sprintf("%d%%", int(math.Round(v*100))) }

func times(v float64) string { return fmt.Sprintf("%.1fx", v) }

func degrees(rad float64) string { return fmt.Sprintf("%d", int(math.Round(rad*180/math.Pi))) }

func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}
