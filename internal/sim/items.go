package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	// healthPickup, armorPickup and the ammo amounts are how much each collectible
	// restores.
	healthPickup = 25.0
	armorPickup  = 25.0
	bulletPickup = 20
	shellPickup  = 8
	rocketPickup = 5
	// noticeDuration is how long a pickup or key message stays on screen, seconds.
	noticeDuration = 2.5
)

// ItemState is a level's collectible plus whether it has been taken. The renderer
// billboards the ones still present; the simulation flips Taken on pickup.
type ItemState struct {
	Kind  world.ItemKind
	Pos   Vec2
	Taken bool
}

// newItems builds the runtime collectible list from a level's layout, centring
// each item on its tile.
func newItems(l *world.Level) []ItemState {
	items := make([]ItemState, len(l.Items))
	for i, it := range l.Items {
		items[i] = ItemState{
			Kind: it.Kind,
			Pos:  Vec2{X: float64(it.At.X) + 0.5, Y: float64(it.At.Y) + 0.5},
		}
	}
	return items
}

// pickupItems collects any item sharing the player's current cell, applying its
// effect, flagging it taken and announcing it.
func (g *Game) pickupItems() {
	pc := g.PlayerCell()
	for i := range g.Items {
		it := &g.Items[i]
		if it.Taken {
			continue
		}
		if int(math.Floor(it.Pos.X)) != pc.X || int(math.Floor(it.Pos.Y)) != pc.Y {
			continue
		}
		g.applyPickup(it.Kind)
		it.Taken = true
		g.items++
		g.emit(Observation{Kind: ObsItem, At: pc})
	}
}

// applyPickup grants the effect of a collected item and sets the pickup notice.
func (g *Game) applyPickup(k world.ItemKind) {
	switch k {
	case world.ItemHealth:
		g.Player.Health = math.Min(g.Player.Health+healthPickup, MaxHealth)
	case world.ItemArmor:
		g.Player.Armor = math.Min(g.Player.Armor+armorPickup, MaxArmor)
	case world.ItemBullets:
		g.Player.Bullets = min(g.Player.Bullets+bulletPickup, g.Player.MaxBullets())
	case world.ItemShells:
		g.Player.Shells = min(g.Player.Shells+shellPickup, g.Player.MaxShells())
	case world.ItemRockets:
		g.Player.Rockets = min(g.Player.Rockets+rocketPickup, g.Player.MaxRockets())
	case world.ItemBackpack:
		g.Player.Backpack = true
		// The pack itself tops up a little of every ammo, up to the new caps.
		g.Player.Bullets = min(g.Player.Bullets+bulletPickup, g.Player.MaxBullets())
		g.Player.Shells = min(g.Player.Shells+shellPickup, g.Player.MaxShells())
		g.Player.Rockets = min(g.Player.Rockets+1, g.Player.MaxRockets())
	default: // keycards
		if g.Player.Keys == nil {
			g.Player.Keys = make(map[world.ItemKind]bool)
		}
		g.Player.Keys[k] = true
	}
	g.setNotice("Picked up " + k.String())
}

// hurtPlayer applies damage to the player, letting armour soak a share of it
// first and clamping health at zero.
func (g *Game) hurtPlayer(dmg float64) {
	if dmg <= 0 {
		return
	}
	absorbed := math.Min(dmg*armorAbsorb, g.Player.Armor)
	g.Player.Armor -= absorbed
	g.Player.Health -= dmg - absorbed
	if g.Player.Health < 0 {
		g.Player.Health = 0
	}
}

// setNotice posts a short-lived on-screen message (pickups, locked doors, finds).
func (g *Game) setNotice(msg string) {
	g.notice = msg
	g.noticeTTL = noticeDuration
}

// Notice returns the current transient message, or "" if none is showing.
func (g *Game) Notice() string {
	if g.noticeTTL <= 0 {
		return ""
	}
	return g.notice
}
