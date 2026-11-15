package sim

type Loadout struct {
	Health, Armor            float64
	Bullets, Shells, Rockets int
	Backpack, Berserk        bool
	Weapon                   WeaponKind
}

func startingLoadout() Loadout {
	return Loadout{Health: MaxHealth, Weapon: Pistol, Bullets: 50, Shells: 20}
}

func (g *Game) Loadout() Loadout {
	p := g.Player
	return Loadout{
		// Patched up between floors: wounds do not follow the player down, but
		// anything above full health does.
		Health:   max(p.Health, MaxHealth),
		Armor:    p.Armor,
		Bullets:  p.Bullets,
		Shells:   p.Shells,
		Rockets:  p.Rockets,
		Backpack: p.Backpack,
		Berserk:  p.Berserk,
		Weapon:   p.Weapon,
	}
}

func (p *Player) equip(l Loadout) {
	p.Health, p.Armor = l.Health, l.Armor
	p.Bullets, p.Shells, p.Rockets = l.Bullets, l.Shells, l.Rockets
	p.Backpack, p.Berserk = l.Backpack, l.Berserk
	p.Weapon = l.Weapon
	p.InvulnTTL, p.RadSuitTTL = 0, 0
}

func WithLoadout(l Loadout) Option {
	return func(g *Game) { g.entry = l }
}

func WithDepth(depth int) Option {
	return func(g *Game) { g.depth = max(depth, 0) }
}
