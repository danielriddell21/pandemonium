package gui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/danielriddell21/crucible/menu"
)

var menuBG = color.RGBA{R: 16, G: 12, B: 14, A: 255}

func label(s string) func() string { return func() string { return s } }

func adjustRow(name string, value func() string, adjust func(int)) menu.Item {
	return menu.Item{
		Label:  func() string { return fmt.Sprintf("%-18s%s", name, value()) },
		Adjust: adjust,
	}
}

func (n nav) input() menu.Input {
	return menu.Input{Up: n.up, Down: n.down, Left: n.left, Right: n.right, Select: n.enter}
}

func percent(v float64) string   { return fmt.Sprintf("%d%%", int(math.Round(v*100))) }
func times(v float64) string     { return fmt.Sprintf("%.1fx", v) }
func degrees(rad float64) string { return fmt.Sprintf("%d", int(math.Round(rad*180/math.Pi))) }

func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}
