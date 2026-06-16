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
	flash          int     // muzzle-flash frames remaining
	viewZ          float64 // camera height easing toward the player's feet

	notice    string  // transient on-screen message (pickups, keys, finds)
	noticeTTL float64 // remaining display time for notice, in seconds

	secrets map[world.Coord]bool // secret cells not yet discovered
	visited map[world.Coord]bool // tiles the player has stepped on (for the automap)

	elapsed                            float64 // seconds simulated this level
	kills, items, found                int     // tallies for this level
	killsTotal, itemsTotal, foundTotal int     // their level-wide totals

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
			Z:       l.Floor(l.Spawn.X, l.Spawn.Y),
			Angle:   facing(l.Spawn, l.Exit),
			Health:  MaxHealth,
			Weapon:  Pistol,
			Bullets: 50,
			Shells:  20,
		},
		Entities: spawnEntities(l),
		Items:    newItems(l),
		secrets:  newSecrets(l),
		visited:  map[world.Coord]bool{l.Spawn: true},
		observer: nopObserver{},
		tracker:  newTracker(l),
	}
	g.viewZ = g.Player.Z
	g.killsTotal = len(g.Entities)
	g.itemsTotal = len(g.Items)
	g.foundTotal = len(l.Secrets)
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Tick advances the simulation by dt seconds given the player's input.
func (g *Game) Tick(in Input, dt float64) {
	g.tick++
	g.elapsed += dt
	g.World.Tick(dt)

	g.Player.Angle = normalizeAngle(g.Player.Angle + in.Turn*turnSpeed*dt + in.TurnDelta)

	dir := g.Player.Dir()
	// Strafe axis is the facing direction rotated 90 degrees.
	strafeX, strafeY := -dir.Y, dir.X
	dx := (dir.X*in.Forward + strafeX*in.Strafe) * moveSpeed * dt
	dy := (dir.Y*in.Forward + strafeY*in.Strafe) * moveSpeed * dt
	g.Player.Pos = resolveMove(g.World, g.Player.Pos, g.Player.Z, dx, dy)
	g.settleHeight(dt)

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

// Visited reports the set of tiles the player has stepped on, for the automap.
// The returned map is owned by the game and must not be mutated by callers.
func (g *Game) Visited() map[world.Coord]bool { return g.visited }

// settleHeight resolves the player's height against the floor underfoot: small
// rises are climbed instantly (and lifts push the body up with the platform),
// while drops fall at a fixed rate. The camera height eases after the body so
// stairs read as steps rather than jolts.
func (g *Game) settleHeight(dt float64) {
	c := g.PlayerCell()
	floor := g.World.FloorAt(c.X, c.Y)
	switch {
	case g.Player.Z < floor:
		g.Player.Z = floor
	case g.Player.Z > floor:
		g.Player.Z = math.Max(floor, g.Player.Z-fallSpeed*dt)
	}

	diff := g.Player.Z - g.viewZ
	step := viewRate * dt
	switch {
	case math.Abs(diff) <= step:
		g.viewZ = g.Player.Z
	case diff > 0:
		g.viewZ += step
	default:
		g.viewZ -= step
	}
}

// EyeZ returns the camera height in wall units: the eased body height plus the
// fixed eye offset. The renderer projects everything relative to this.
func (g *Game) EyeZ() float64 { return g.viewZ + eyeHeight }

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
	g.visited[cell] = true

	g.emit(Observation{Kind: ObsMove, At: cell})
	g.checkSecret(cell)

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
	if !g.World.IsDoor(tx, ty) || g.World.Opened(tx, ty) {
		return
	}
	if key, locked := g.World.Lock(tx, ty); locked && !g.Player.HasKey(key) {
		g.setNotice("You need the " + key.String())
		return
	}
	if !g.World.OpenDoor(tx, ty, g.Player.HasKey) {
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
