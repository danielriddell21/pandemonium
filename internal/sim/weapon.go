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
)

type ammoKind uint8

const (
	ammoNone ammoKind = iota
	ammoBullets
	ammoShells
)

// weaponSpec describes how a weapon fires.
type weaponSpec struct {
	rng      float64 // reach in tiles
	arcCos   float64 // cosine of the half-angle it can hit within
	damage   float64 // damage per pellet/shot
	targets  int     // how many demons a single shot can hit (spread)
	ammo     ammoKind
	cooldown float64 // seconds between shots
}

var weapons = map[WeaponKind]weaponSpec{
	Fists:   {rng: 1.6, arcCos: 0.80, damage: 50, targets: 1, ammo: ammoNone, cooldown: 0.40},
	Pistol:  {rng: 9.0, arcCos: 0.97, damage: 28, targets: 1, ammo: ammoBullets, cooldown: 0.45},
	Shotgun: {rng: 7.0, arcCos: 0.82, damage: 22, targets: 4, ammo: ammoShells, cooldown: 0.80},
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
	for _, i := range g.hitscanMulti(w.rng, w.arcCos, w.targets) {
		g.damageEntity(i, w.damage)
	}
	return true
}

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
	}
	return true
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

// switchWeapon selects a weapon by slot (1=fists, 2=pistol, 3=shotgun).
func (g *Game) switchWeapon(sel int) {
	switch sel {
	case 1:
		g.Player.Weapon = Fists
	case 2:
		g.Player.Weapon = Pistol
	case 3:
		g.Player.Weapon = Shotgun
	}
}

// MuzzleFlash reports whether the weapon fired recently enough to show a flash.
func (g *Game) MuzzleFlash() bool {
	return g.flash > 0
}
