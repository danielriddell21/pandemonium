package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	// moveSpeed is the player's translation speed in tiles per second.
	moveSpeed = 3.2
	// turnSpeed is the player's rotation speed in radians per second when
	// turning via keyboard. Mouse turning is applied directly as a delta.
	turnSpeed = 2.6
	// reach is how far ahead the player can act on a door, in tiles.
	reach = 0.9
)

// Game is the full simulation state for one level: the world, the player and the
// entities within it. It advances via Tick and owns no rendering.
type Game struct {
	World       *World
	Player      Player
	Entities    []Entity
	Projectiles []Projectile
	Items       []ItemState
	tick        uint64

	attackCooldown float64
	flash          int // muzzle-flash frames remaining

	notice    string  // transient on-screen message (pickups, keys, finds)
	noticeTTL float64 // remaining display time for notice, in seconds

	observer Observer
	tracker  tracker
}

// New builds a simulation for a generated level, placing the player at the spawn
// facing roughly toward the exit and scattering demons across the map. Options
// may attach an observer; by default observations are discarded.
func New(l *world.Level, opts ...Option) *Game {
	g := &Game{
		World: NewWorld(l),
		Player: Player{
			Pos:     Vec2{X: float64(l.Spawn.X) + 0.5, Y: float64(l.Spawn.Y) + 0.5},
			Angle:   facing(l.Spawn, l.Exit),
			Health:  MaxHealth,
			Weapon:  Pistol,
			Bullets: 50,
			Shells:  20,
		},
		Entities: spawnEntities(l),
		Items:    newItems(l),
		observer: nopObserver{},
		tracker:  newTracker(l),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Tick advances the simulation by dt seconds given the player's input.
func (g *Game) Tick(in Input, dt float64) {
	g.tick++

	g.Player.Angle = normalizeAngle(g.Player.Angle + in.Turn*turnSpeed*dt + in.TurnDelta)

	dir := g.Player.Dir()
	// Strafe axis is the facing direction rotated 90 degrees.
	strafeX, strafeY := -dir.Y, dir.X
	dx := (dir.X*in.Forward + strafeX*in.Strafe) * moveSpeed * dt
	dy := (dir.Y*in.Forward + strafeY*in.Strafe) * moveSpeed * dt
	g.Player.Pos = resolveMove(g.World, g.Player.Pos, dx, dy)

	if in.SelectWeapon != 0 {
		g.switchWeapon(in.SelectWeapon)
	}

	g.updateEntities(dt)

	if g.attackCooldown > 0 {
		g.attackCooldown -= dt
	}
	if g.flash > 0 {
		g.flash--
	}
	if in.Attack && g.attackCooldown <= 0 {
		if g.fire() {
			g.attackCooldown = weapons[g.Player.Weapon].cooldown
		}
	}

	g.advanceProjectiles(dt)
	g.applyContactDamage(dt)
	if g.Player.Health <= 0 {
		g.die()
	}

	g.pickupItems()
	if g.noticeTTL > 0 {
		g.noticeTTL -= dt
	}

	g.observeMovement()

	if in.Interact {
		g.interact()
	}

	if g.ReachedExit() && !g.tracker.exitEmitted {
		g.tracker.exitEmitted = true
		g.emit(Observation{Kind: ObsExit, At: g.PlayerCell()})
	}
}

// emit stamps the current tick onto an observation and hands it to the observer.
func (g *Game) emit(o Observation) {
	o.Tick = g.tick
	g.observer.Observe(o)
}

// observeMovement emits an observation whenever the player enters a new tile,
// plus a richer one when that tile carries a marker.
func (g *Game) observeMovement() {
	cell := g.PlayerCell()
	if g.tracker.started && cell == g.tracker.lastCell {
		return
	}
	g.tracker.started = true
	g.tracker.lastCell = cell

	g.emit(Observation{Kind: ObsMove, At: cell})

	mk, ok := g.tracker.markerAt(cell)
	if !ok {
		return
	}
	obs := Observation{Kind: ObsMarker, At: cell, Marker: mk.Kind}
	if mk.Kind == world.MarkerJunction {
		obs.Taken, obs.Ignored = splitBranches(mk, g.Player.Dir(), cell)
		obs.Optimal = mk.Optimal
	}
	g.emit(obs)
}

// Tick64 returns the number of ticks simulated so far, for pacing queries.
func (g *Game) Tick64() uint64 { return g.tick }

// PlayerCell returns the integer tile the player currently occupies.
func (g *Game) PlayerCell() world.Coord {
	return world.Coord{X: int(math.Floor(g.Player.Pos.X)), Y: int(math.Floor(g.Player.Pos.Y))}
}

// ReachedExit reports whether the player is standing on the exit tile.
func (g *Game) ReachedExit() bool {
	c := g.PlayerCell()
	return g.World.Level.At(c.X, c.Y) == world.TileExit
}

// interact opens a door immediately ahead of the player, if any, and reports it.
func (g *Game) interact() {
	dir := g.Player.Dir()
	tx := int(math.Floor(g.Player.Pos.X + dir.X*reach))
	ty := int(math.Floor(g.Player.Pos.Y + dir.Y*reach))
	if !g.World.OpenDoor(tx, ty) {
		return
	}
	cell := world.Coord{X: tx, Y: ty}
	wrong := false
	if mk, ok := g.tracker.markerAt(cell); ok {
		wrong = mk.Kind == world.MarkerDeadEndDoor
	}
	g.emit(Observation{Kind: ObsDoor, At: cell, WrongDoor: wrong})
}

// facing returns the angle pointing from a toward b, used to orient the player
// toward the exit at spawn.
func facing(a, b world.Coord) float64 {
	return math.Atan2(float64(b.Y-a.Y), float64(b.X-a.X))
}
