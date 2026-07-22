package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	moveSpeed = 3.2

	turnSpeed = 2.6

	reach = 0.9
)

type Game struct {
	World       *World
	Player      Player
	Entities    []Entity
	Projectiles []Projectile
	Items       []ItemState
	tick        uint64

	attackCooldown float64
	flash          int
	viewZ          float64

	notice    string
	noticeTTL float64
	exitDone  bool

	secrets map[world.Coord]bool
	visited map[world.Coord]bool

	elapsed                            float64
	par                                float64
	kills, items, found                int
	killsTotal, itemsTotal, foundTotal int

	observer Observer
	tracker  tracker
	skill    Skill
}

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
		Items:    newItems(l),
		secrets:  newSecrets(l),
		visited:  map[world.Coord]bool{l.Spawn: true},
		observer: nopObserver{},
		tracker:  newTracker(l),
		skill:    SkillNormal,
	}
	// Apply options before spawning so the difficulty governs the encounter.
	for _, opt := range opts {
		opt(g)
	}
	g.Entities = g.spawnEntities()
	g.viewZ = g.Player.Z
	g.killsTotal = countDemons(g.Entities)
	g.itemsTotal = len(g.Items)
	g.foundTotal = len(l.Secrets)
	g.par = parTime(world.StepsBetween(l, l.Spawn, l.Exit))
	return g
}

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
	g.autoSwitchIfEmpty()

	g.advanceProjectiles(dt)
	g.applyContactDamage(dt)
	g.applyHazard(dt)
	if g.Player.Health <= 0 {
		g.die()
	}

	g.pickupItems()
	g.tickPowerups(dt)
	if g.noticeTTL > 0 {
		g.noticeTTL -= dt
	}

	g.observeMovement()

	if in.Interact {
		g.interact()
	}

	if g.LevelComplete() && !g.tracker.exitEmitted {
		g.tracker.exitEmitted = true
		g.emit(Observation{Kind: ObsExit, At: g.PlayerCell()})
	}
}

func (g *Game) Visited() map[world.Coord]bool { return g.visited }

func (g *Game) settleHeight(dt float64) {
	c := g.PlayerCell()
	floor := g.World.FloorAt(c.X, c.Y)
	switch {
	case g.Player.Z < floor:
		g.Player.Z = floor
	case g.Player.Z > floor:
		g.Player.Z = max(floor, g.Player.Z-fallSpeed*dt)
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

func (g *Game) EyeZ() float64 { return g.viewZ + eyeHeight }

func (g *Game) emit(o Observation) {
	o.Tick = g.tick
	g.observer.Observe(o)
}

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

func (g *Game) Tick64() uint64 { return g.tick }

func (g *Game) PlayerCell() world.Coord {
	return world.Coord{X: int(math.Floor(g.Player.Pos.X)), Y: int(math.Floor(g.Player.Pos.Y))}
}

func (g *Game) ReachedExit() bool {
	c := g.PlayerCell()
	return g.World.Level.At(c.X, c.Y) == world.TileExit
}

func (g *Game) LevelComplete() bool {
	if g.World.Level.HasExitSwitch() {
		return g.exitDone
	}
	return g.ReachedExit()
}

func (g *Game) pressSwitch(sw world.Switch) {
	switch sw.Action {
	case world.SwitchExit:
		g.exitDone = true
	case world.SwitchDoor:
		g.World.ForceOpenDoor(sw.Target.X, sw.Target.Y)
	}
}

func (g *Game) interact() {
	dir := g.Player.Dir()
	tx := int(math.Floor(g.Player.Pos.X + dir.X*reach))
	ty := int(math.Floor(g.Player.Pos.Y + dir.Y*reach))

	if sw, ok := g.World.Level.SwitchAt(tx, ty); ok {
		g.pressSwitch(sw)
		return
	}

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

func facing(a, b world.Coord) float64 {
	return math.Atan2(float64(b.Y-a.Y), float64(b.X-a.X))
}
