package sim

import (
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	alarmDecay    = 0.015
	alarmOvertime = 0.02
	alarmQuiet    = 25.0
	alarmLull     = 12.0
	alarmDepth    = 0.05
	alarmMaxGain  = 1.5

	blastNoise = 0.1

	waveBase    = 2
	waveMax     = 8
	waveMinDist = 8
	waveSalt    = 0x9e3779b97f4a7c15
)

var weaponNoise = map[WeaponKind]float64{
	Fists:          0,
	Pistol:         0.04,
	Shotgun:        0.08,
	Chaingun:       0.022,
	RocketLauncher: 0.03,
}

var waveKinds = [...]EntityKind{Melee, Pinky, Ranged, Melee, Gunner, Pinky, Ranged, Baron}

var spoils = [...]world.ItemKind{
	world.ItemBullets, world.ItemShells, world.ItemHealth,
	world.ItemBullets, world.ItemRockets, world.ItemArmor,
}

type alarm struct {
	meter     float64
	waves     int
	quiet     float64
	flash     float64
	panelUsed bool
	drops     int
}

const alarmFlashTime = 1.5

func (g *Game) AlarmFlash() float64 { return max(g.alarm.flash, 0) / alarmFlashTime }

func (g *Game) Alert() float64 { return min(g.alarm.meter, 1) }

func (g *Game) Alarms() int { return g.alarm.waves }

func (g *Game) AlarmQuiet() bool { return g.alarm.panelUsed && g.alarm.quiet > 0 }

func (g *Game) alarmGain() float64 {
	return min(1+alarmDepth*float64(g.depth), alarmMaxGain)
}

func (g *Game) alarmArmed() bool {
	return g.alarm.quiet <= 0 && !g.World.Level.Boss
}

func (g *Game) noise(n float64) {
	if g.alarmArmed() {
		g.alarm.meter += n * g.alarmGain()
	}
}

func (g *Game) tickAlarm(dt float64) {
	a := &g.alarm
	if a.quiet > 0 {
		a.quiet -= dt
	}
	if a.flash > 0 {
		a.flash -= dt
	}
	if !g.alarmArmed() {
		return
	}
	// A meter the last shot filled sounds before it can settle back under.
	if a.meter >= 1 {
		g.raiseAlarm()
		return
	}
	// Under par the level settles between shots; past it, lingering alone is
	// enough to bring the horde.
	if g.elapsed > g.par {
		a.meter += alarmOvertime * g.alarmGain() * dt
	} else {
		a.meter = max(0, a.meter-alarmDecay*dt)
	}
}

func (g *Game) raiseAlarm() {
	a := &g.alarm
	a.meter = 0
	a.waves++
	// The klaxon is already sounding: the fight it starts cannot at once set
	// off the next one.
	a.quiet = alarmLull
	a.flash = alarmFlashTime
	for i := range g.Entities {
		if e := &g.Entities[i]; e.Alive && e.Kind != Barrel {
			e.alert = true
		}
	}
	g.spawnWave(g.waveSize(), g.waveCells())
	g.setNotice("ALARM! Reinforcements inbound")
	g.emit(Observation{Kind: ObsAlarm, At: g.PlayerCell()})
}

func (g *Game) waveSize() int {
	n := float64(waveBase+g.alarm.waves+g.depth/3) * g.skill.countScale()
	return min(max(int(n), 1), waveMax)
}

func (g *Game) waveCells() []world.Coord {
	l := g.World.Level
	var hidden, far []world.Coord
	for y := range l.H {
		for x := range l.W {
			c := world.Coord{X: x, Y: y}
			if l.At(x, y) != world.TileFloor || g.scentAt(c) < waveMinDist {
				continue
			}
			far = append(far, c)
			if !losClear(g.World, centreOf(c), g.Player.Pos) {
				hidden = append(hidden, c)
			}
		}
	}
	// Reinforcements arrive out of sight when the level allows it.
	if len(hidden) > 0 {
		far = hidden
	}
	r := rand.New(rand.NewPCG(uint64(l.Seed), waveSalt*uint64(g.alarm.waves)))
	r.Shuffle(len(far), func(i, j int) { far[i], far[j] = far[j], far[i] })
	return far
}

func (g *Game) spawnWave(n int, cells []world.Coord) {
	for i := range min(n, len(cells)) {
		kind := waveKinds[i%len(waveKinds)]
		if kind == Baron && g.alarm.waves < 3 {
			kind = Melee // the heavy hitters only answer a level that keeps ringing
		}
		g.spawnMinion(cells[i], kind)
	}
}

func (g *Game) spawnMinion(c world.Coord, kind EntityKind) {
	d := newDemon(centreOf(c), kind)
	d.Z = g.World.FloorAt(c.X, c.Y)
	d.alert, d.wave = true, true
	g.Entities = append(g.Entities, d)
}

func (g *Game) dropSpoils(e *Entity) {
	kind := spoils[g.alarm.drops%len(spoils)]
	g.alarm.drops++
	g.Items = append(g.Items, ItemState{Kind: kind, Pos: e.Pos, Bonus: true})
}

func (g *Game) silence() {
	a := &g.alarm
	if a.panelUsed {
		g.setNotice("The panel is dead")
		return
	}
	a.panelUsed = true
	a.meter = 0
	a.quiet = alarmQuiet
	// The hunt loses the scent; anything still in sight simply picks it up again.
	for i := range g.Entities {
		if g.Entities[i].Kind != Boss {
			g.Entities[i].alert = false
		}
	}
	g.setNotice("Alarm silenced")
	g.emit(Observation{Kind: ObsSilence, At: g.PlayerCell()})
}

func (g *Game) PanelSpent() bool { return g.alarm.panelUsed }

func hitRadius(k EntityKind) float64 {
	if k == Boss {
		return bossRadius
	}
	return projectileHitRadius
}

func angleOf(v Vec2) float64 { return math.Atan2(v.Y, v.X) }
