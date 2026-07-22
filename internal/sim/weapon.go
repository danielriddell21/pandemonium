package sim

import (
	"cmp"
	"math"
	"slices"
)

type WeaponKind uint8

const (
	Fists WeaponKind = iota

	Pistol

	Shotgun

	Chaingun

	RocketLauncher
)

type ammoKind uint8

const (
	ammoNone ammoKind = iota
	ammoBullets
	ammoShells
	ammoRockets
)

type weaponSpec struct {
	rng      float64
	arcCos   float64
	damage   float64
	targets  int
	ammo     ammoKind
	cooldown float64
	rocket   bool
}

var weapons = map[WeaponKind]weaponSpec{
	Fists:          {rng: 1.6, arcCos: 0.80, damage: 50, targets: 1, ammo: ammoNone, cooldown: 0.40},
	Pistol:         {rng: 9.0, arcCos: 0.97, damage: 28, targets: 1, ammo: ammoBullets, cooldown: 0.45},
	Shotgun:        {rng: 7.0, arcCos: 0.82, damage: 22, targets: 4, ammo: ammoShells, cooldown: 0.80},
	Chaingun:       {rng: 9.0, arcCos: 0.97, damage: 18, targets: 1, ammo: ammoBullets, cooldown: 0.12},
	RocketLauncher: {damage: 70, ammo: ammoRockets, cooldown: 0.85, rocket: true},
}

const muzzleFlashTicks = 5

func (g *Game) fire() bool {
	w := weapons[g.Player.Weapon]
	if !g.spendAmmo(w.ammo) {
		return false
	}
	g.flash = muzzleFlashTicks
	if w.rocket {
		g.spawnPlayerRocket(w.damage)
		return true
	}
	dmg := w.damage
	if g.Player.Weapon == Fists && g.Player.Berserk {
		dmg = berserkFistDamage // berserk turns the fists into a one-punch kill
	}
	for _, i := range g.hitscanMulti(w.rng, w.arcCos, w.targets) {
		g.damageEntity(i, dmg)
	}
	return true
}

const berserkFistDamage = 200.0

func (g *Game) spendAmmo(a ammoKind) bool {
	switch a {
	case ammoBullets:
		if g.Player.Bullets <= 0 {
			return false
		}
		g.Player.Bullets--
	case ammoShells:
		if g.Player.Shells <= 0 {
			return false
		}
		g.Player.Shells--
	case ammoRockets:
		if g.Player.Rockets <= 0 {
			return false
		}
		g.Player.Rockets--
	}
	return true
}

func (g *Game) ammoCount(a ammoKind) int {
	switch a {
	case ammoBullets:
		return g.Player.Bullets
	case ammoShells:
		return g.Player.Shells
	case ammoRockets:
		return g.Player.Rockets
	default:
		return 1
	}
}

func (g *Game) autoSwitchIfEmpty() {
	w := weapons[g.Player.Weapon]
	if w.ammo == ammoNone || g.ammoCount(w.ammo) > 0 {
		return
	}
	for _, k := range []WeaponKind{Chaingun, Shotgun, Pistol, Fists} {
		if k != g.Player.Weapon && g.ammoCount(weapons[k].ammo) > 0 {
			g.Player.Weapon = k
			return
		}
	}
}

func (g *Game) hitscanMulti(maxRange, arcCos float64, n int) []int {
	type cand struct {
		i int
		d float64
	}
	cs := make([]cand, 0, len(g.Entities))
	dir := g.Player.Dir()
	for i := range g.Entities {
		e := g.Entities[i]
		if !e.Alive {
			continue
		}
		dx, dy := e.Pos.X-g.Player.Pos.X, e.Pos.Y-g.Player.Pos.Y
		d := math.Hypot(dx, dy)
		if d == 0 || d > maxRange {
			continue
		}
		if (dx/d)*dir.X+(dy/d)*dir.Y < arcCos {
			continue
		}
		if !losClear(g.World, g.Player.Pos, e.Pos) {
			continue
		}
		cs = append(cs, cand{i, d})
	}
	slices.SortFunc(cs, func(a, b cand) int { return cmp.Compare(a.d, b.d) })
	if len(cs) > n {
		cs = cs[:n]
	}
	out := make([]int, len(cs))
	for k, c := range cs {
		out[k] = c.i
	}
	return out
}

func (g *Game) switchWeapon(sel int) {
	switch sel {
	case 1:
		g.Player.Weapon = Fists
	case 2:
		g.Player.Weapon = Pistol
	case 3:
		g.Player.Weapon = Shotgun
	case 4:
		g.Player.Weapon = Chaingun
	case 5:
		g.Player.Weapon = RocketLauncher
	}
}

func (g *Game) MuzzleFlash() bool {
	return g.flash > 0
}
