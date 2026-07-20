package sim

import (
	"math"
	"sort"
)

// WeaponKind identifies the player's selectable weapons.
type WeaponKind uint8

const (
	// Fists are the melee fallback; short reach, no ammo.
	Fists WeaponKind = iota
	// Pistol is a precise single-target hitscan.
	Pistol
	// Shotgun sprays several pellets across a wide arc.
	Shotgun
	// Chaingun is a rapid single-target hitscan that chews through bullets.
	Chaingun
	// RocketLauncher fires a slow rocket that bursts for splash damage on impact.
	RocketLauncher
)

type ammoKind uint8

const (
	ammoNone ammoKind = iota
	ammoBullets
	ammoShells
	ammoRockets
)

// weaponSpec describes how a weapon fires.
type weaponSpec struct {
	rng      float64 // reach in tiles (hitscan weapons)
	arcCos   float64 // cosine of the half-angle it can hit within
	damage   float64 // damage per pellet/shot (or splash damage for rockets)
	targets  int     // how many demons a single shot can hit (spread)
	ammo     ammoKind
	cooldown float64 // seconds between shots
	rocket   bool    // fires a splash projectile instead of a hitscan
}

var weapons = map[WeaponKind]weaponSpec{
	Fists:          {rng: 1.6, arcCos: 0.80, damage: 50, targets: 1, ammo: ammoNone, cooldown: 0.40},
	Pistol:         {rng: 9.0, arcCos: 0.97, damage: 28, targets: 1, ammo: ammoBullets, cooldown: 0.45},
	Shotgun:        {rng: 7.0, arcCos: 0.82, damage: 22, targets: 4, ammo: ammoShells, cooldown: 0.80},
	Chaingun:       {rng: 9.0, arcCos: 0.97, damage: 18, targets: 1, ammo: ammoBullets, cooldown: 0.12},
	RocketLauncher: {damage: 70, ammo: ammoRockets, cooldown: 0.85, rocket: true},
}

// muzzleFlashTicks is how many ticks the muzzle flash shows after firing.
const muzzleFlashTicks = 5

// fire discharges the current weapon: it spends ammo, records the shot for the
// muzzle flash, and wounds up to the weapon's target count of demons ahead. It
// reports whether a shot actually went off (false when out of ammo).
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

// berserkFistDamage is the punishing fist damage while berserk is active.
const berserkFistDamage = 200.0

// spendAmmo consumes one round of the given kind, reporting success.
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

// ammoCount returns how many rounds of the given kind the player holds. Weapons
// that need no ammo (the fists) always report a usable count.
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

// autoSwitchIfEmpty drops to the best still-usable weapon when the current one
// runs dry, the way DOOM falls back after you fire your last round. The rocket
// launcher is never auto-selected (its ammo is too precious to spend by reflex),
// so the fists are the final fallback.
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

// hitscanMulti returns up to n nearest living demons within range, inside the
// facing arc, and in clear line of sight, nearest first.
func (g *Game) hitscanMulti(maxRange, arcCos float64, n int) []int {
	type cand struct {
		i int
		d float64
	}
	var cs []cand
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
	sort.Slice(cs, func(a, b int) bool { return cs[a].d < cs[b].d })
	if len(cs) > n {
		cs = cs[:n]
	}
	out := make([]int, len(cs))
	for k, c := range cs {
		out[k] = c.i
	}
	return out
}

// switchWeapon selects a weapon by slot (1=fists, 2=pistol, 3=shotgun,
// 4=chaingun, 5=rocket launcher).
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

// MuzzleFlash reports whether the weapon fired recently enough to show a flash.
func (g *Game) MuzzleFlash() bool {
	return g.flash > 0
}
