package sim

import "testing"

func TestVisitedGrowsAsPlayerMoves(t *testing.T) {
	g := newTestGame(t, 7)
	start := len(g.Visited())
	if start == 0 {
		t.Fatal("spawn tile should start visited")
	}
	for range 240 {
		g.Tick(Input{Forward: 1}, 1.0/60.0)
	}
	if len(g.Visited()) <= start {
		t.Errorf("visited set did not grow as the player moved: %d -> %d", start, len(g.Visited()))
	}
	// Every visited cell must be the player's current or a previously entered tile;
	// the current cell is always present.
	if !g.Visited()[g.PlayerCell()] {
		t.Error("current player cell not recorded as visited")
	}
}
