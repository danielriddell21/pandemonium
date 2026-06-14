package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestLevelStatsTotalsFromLayout(t *testing.T) {
	g := newTestGame(t, 7)
	s := g.LevelStats()
	if s.KillsTotal != len(g.Entities) {
		t.Errorf("KillsTotal = %d, want %d", s.KillsTotal, len(g.Entities))
	}
	if s.ItemsTotal != len(g.Items) {
		t.Errorf("ItemsTotal = %d, want %d", s.ItemsTotal, len(g.Items))
	}
	if s.Kills != 0 || s.Items != 0 || s.Secrets != 0 {
		t.Errorf("fresh level should have zero tallies, got %+v", s)
	}
}

func TestKillTallyIncrementsOnDeath(t *testing.T) {
	g := newTestGame(t, 7)
	g.Entities = []Entity{{Pos: Vec2{X: g.Player.Pos.X + 1.5, Y: g.Player.Pos.Y}, State: Active, Health: 10, Alive: true}}
	g.killsTotal = 1
	g.damageEntity(0, 50) // lethal

	s := g.LevelStats()
	if s.Kills != 1 {
		t.Errorf("Kills = %d, want 1", s.Kills)
	}
	if s.KillsPct() != 100 {
		t.Errorf("KillsPct = %d, want 100", s.KillsPct())
	}
}

func TestItemTallyIncrementsOnPickup(t *testing.T) {
	g := newTestGame(t, 7)
	g.Items = []ItemState{{Kind: world.ItemHealth, Pos: g.Player.Pos}}
	g.itemsTotal = 1
	g.Tick(Input{}, 1.0/60.0)
	if g.LevelStats().Items != 1 {
		t.Errorf("Items = %d, want 1", g.LevelStats().Items)
	}
}

func TestStatsPctHandlesEmpty(t *testing.T) {
	var s LevelStats // all zero, nothing to find
	if s.KillsPct() != 100 || s.ItemsPct() != 100 || s.SecretsPct() != 100 {
		t.Errorf("empty stats should read 100%%, got %d/%d/%d", s.KillsPct(), s.ItemsPct(), s.SecretsPct())
	}
}
