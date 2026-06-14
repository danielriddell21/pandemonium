package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

func TestIntermissionRendersStats(t *testing.T) {
	cfg := Config{Width: 320, Height: 200, FOV: 1.152}
	fb := make([]byte, cfg.Width*cfg.Height*4)
	stats := sim.LevelStats{Kills: 3, KillsTotal: 4, Items: 2, ItemsTotal: 5, Secrets: 1, SecretsTotal: 1, Elapsed: 95}

	drawIntermission(fb, cfg, stats)

	if countColored(fb, cfg) == 0 {
		t.Error("intermission screen drew nothing")
	}
}

func TestFormatTime(t *testing.T) {
	cases := map[float64]string{0: "0:00", 9: "0:09", 65: "1:05", 600: "10:00"}
	for secs, want := range cases {
		if got := formatTime(secs); got != want {
			t.Errorf("formatTime(%v) = %q, want %q", secs, got, want)
		}
	}
}

func TestNoticeRendersText(t *testing.T) {
	cfg := Config{Width: 320, Height: 200, FOV: 1.152}
	blank := make([]byte, cfg.Width*cfg.Height*4)
	withMsg := make([]byte, cfg.Width*cfg.Height*4)

	drawNotice(blank, cfg, "")
	drawNotice(withMsg, cfg, "Picked up shells")

	if countColored(blank, cfg) != 0 {
		t.Error("empty notice should draw nothing")
	}
	if countColored(withMsg, cfg) == 0 {
		t.Error("notice text drew nothing")
	}
}
