package sim

import "testing"

func TestParScalesWithDistance(t *testing.T) {
	near := parTime(5)
	far := parTime(60)
	if !(far > near && near > 0) {
		t.Errorf("par should grow with distance: near=%v far=%v", near, far)
	}
	if parTime(-1) != parTime(0) {
		t.Error("an unreachable distance should fall back to the base par")
	}
}

func TestLevelStatsCarriesPar(t *testing.T) {
	g := newTestGame(t, 7)
	if g.LevelStats().Par <= 0 {
		t.Error("a generated level should have a positive par time")
	}
}
