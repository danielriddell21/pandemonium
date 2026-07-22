package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	healthPickup = 25.0
	armorPickup  = 25.0
	bulletPickup = 20
	shellPickup  = 8
	rocketPickup = 5

	noticeDuration = 2.5
)

type ItemState struct {
	Kind  world.ItemKind
	Pos   Vec2
	Taken bool
}

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

func (g *Game) applyPickup(k world.ItemKind) {
	switch k {
	case world.ItemHealth:
		g.Player.Health = min(g.Player.Health+healthPickup, MaxHealth)
	case world.ItemArmor:
		g.Player.Armor = min(g.Player.Armor+armorPickup, MaxArmor)
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
	case world.ItemSoul:
		g.Player.Health = min(g.Player.Health+100, overHealMax)
	case world.ItemMega:
		g.Player.Health = overHealMax
		g.Player.Armor = overHealMax
	case world.ItemBerserk:
		g.Player.Berserk = true
		g.Player.Health = max(g.Player.Health, MaxHealth)
	case world.ItemInvuln:
		g.Player.InvulnTTL = invulnDuration
	case world.ItemRadSuit:
		g.Player.RadSuitTTL = radSuitDuration
	default: // keycards
		if g.Player.Keys == nil {
			g.Player.Keys = make(map[world.ItemKind]bool)
		}
		g.Player.Keys[k] = true
	}
	g.setNotice("Picked up " + k.String())
}

func (g *Game) hurtPlayer(dmg float64) {
	if dmg <= 0 || g.Player.Invulnerable() {
		return
	}
	dmg *= g.skill.damageScale()
	absorbed := min(dmg*armorAbsorb, g.Player.Armor)
	g.Player.Armor -= absorbed
	g.Player.Health -= dmg - absorbed
	if g.Player.Health < 0 {
		g.Player.Health = 0
	}
}

func (g *Game) tickPowerups(dt float64) {
	if g.Player.InvulnTTL > 0 {
		g.Player.InvulnTTL -= dt
	}
	if g.Player.RadSuitTTL > 0 {
		g.Player.RadSuitTTL -= dt
	}
	if g.Player.hurtTTL > 0 {
		g.Player.hurtTTL -= dt
	}
}

func (g *Game) setNotice(msg string) {
	g.notice = msg
	g.noticeTTL = noticeDuration
}

func (g *Game) Notice() string {
	if g.noticeTTL <= 0 {
		return ""
	}
	return g.notice
}
