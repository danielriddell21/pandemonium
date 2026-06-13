package render

import "testing"

func countColored(fb []byte, cfg Config) int {
	n := 0
	for i := 0; i+3 < len(fb); i += 4 {
		if fb[i] != 0 || fb[i+1] != 0 || fb[i+2] != 0 {
			n++
		}
	}
	_ = cfg
	return n
}

func TestHealthBarScalesWithHealth(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	full := make([]byte, cfg.Width*cfg.Height*4)
	half := make([]byte, len(full))
	empty := make([]byte, len(full))

	drawHealthBar(full, cfg, 1.0)
	drawHealthBar(half, cfg, 0.5)
	drawHealthBar(empty, cfg, 0.0)

	cf, ch, ce := countColored(full, cfg), countColored(half, cfg), countColored(empty, cfg)
	if !(cf > ch && ch > ce) {
		t.Errorf("expected full > half > empty filled pixels; got %d, %d, %d", cf, ch, ce)
	}
	if ce != 0 {
		t.Errorf("empty health should draw no coloured fill, got %d", ce)
	}
}
