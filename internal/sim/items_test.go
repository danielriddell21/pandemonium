package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

// placeItemUnderPlayer drops a single item on the player's current cell so a tick
// will collect it.
func placeItemUnderPlayer(g *Game, kind world.ItemKind) {
	g.Items = []ItemState{{Kind: kind, Pos: g.Player.Pos}}
}

func TestHealthPickupRestoresAndIsTaken(t *testing.T) {
	g := newTestGame(t, 7)
	capt := &captureObserver{}
	g.observer = capt
	g.Player.Health = 40
	placeItemUnderPlayer(g, world.ItemHealth)

	g.Tick(Input{}, 1.0/60.0)

	if g.Player.Health != 40+healthPickup {
		t.Errorf("health = %v, want %v", g.Player.Health, 40+healthPickup)
	}
	if !g.Items[0].Taken {
		t.Error("item should be marked taken after pickup")
	}
	if capt.count(ObsItem) != 1 {
		t.Errorf("ObsItem count = %d, want 1", capt.count(ObsItem))
	}
	if g.Notice() == "" {
		t.Error("expected a pickup notice")
	}
}

func TestHealthPickupClampsAtMax(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Health = MaxHealth - 5
	placeItemUnderPlayer(g, world.ItemHealth)
	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Health != MaxHealth {
		t.Errorf("health = %v, want clamp at %v", g.Player.Health, MaxHealth)
	}
}

func TestAmmoAndArmorPickups(t *testing.T) {
	g := newTestGame(t, 7)
	bullets := g.Player.Bullets
	placeItemUnderPlayer(g, world.ItemBullets)
	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Bullets != bullets+bulletPickup {
		t.Errorf("bullets = %d, want %d", g.Player.Bullets, bullets+bulletPickup)
	}

	placeItemUnderPlayer(g, world.ItemArmor)
	g.Tick(Input{}, 1.0/60.0)
	if g.Player.Armor != armorPickup {
		t.Errorf("armor = %v, want %v", g.Player.Armor, armorPickup)
	}
}

func TestKeyPickupGrantsKey(t *testing.T) {
	g := newTestGame(t, 7)
	placeItemUnderPlayer(g, world.ItemKeyRed)
	g.Tick(Input{}, 1.0/60.0)
	if !g.Player.HasKey(world.ItemKeyRed) {
		t.Error("player should hold the red key after pickup")
	}
}

func TestArmorAbsorbsDamage(t *testing.T) {
	g := newTestGame(t, 7)
	g.Items = nil
	g.Player.Health = MaxHealth
	g.Player.Armor = 60
	g.hurtPlayer(30)

	// Armour soaks a third (10), health loses the remaining 20.
	if got := g.Player.Health; got != MaxHealth-20 {
		t.Errorf("health = %v, want %v", got, MaxHealth-20)
	}
	if got := g.Player.Armor; got != 50 {
		t.Errorf("armor = %v, want 50", got)
	}
}

func TestArmorDepletes(t *testing.T) {
	g := newTestGame(t, 7)
	g.Player.Health = MaxHealth
	g.Player.Armor = 5 // less than the third it would otherwise absorb
	g.hurtPlayer(30)
	if g.Player.Armor != 0 {
		t.Errorf("armor should deplete to 0, got %v", g.Player.Armor)
	}
	if g.Player.Health != MaxHealth-25 {
		t.Errorf("health = %v, want %v (30 - 5 absorbed)", g.Player.Health, MaxHealth-25)
	}
}
