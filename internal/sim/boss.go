package sim

import (
	"math"

	"github.com/danielriddell21/pandemonium/internal/world"
)

const (
	bossHealthBase = 2000.0
	bossHealthStep = 600.0
	bossRadius     = 0.9
	bossSight      = 16.0

	volleyWindup = 0.7
	summonWindup = 1.0
	chargeWindup = 0.9

	volleyDamage = 18.0
	volleySpeed  = 5.5
	volleySpread = 0.2

	chargeSpeed  = 9.0
	chargeTime   = 1.4
	chargeDamage = 45.0
	chargeStall  = 0.5

	stunTime   = 2.5
	stunWeak   = 2.0
	minionCap  = 5
	bossPhases = 3
)

var (
	bossSpeed    = [bossPhases]float64{1.0, 1.3, 1.7}
	bossCooldown = [bossPhases]float64{2.4, 2.0, 1.6}
	bossMoves    = [bossPhases][]BossAction{
		{BossVolley},
		{BossVolley, BossSummon},
		{BossCharge, BossVolley, BossSummon},
	}
)

type BossAction uint8

const (
	BossStalk BossAction = iota

	BossVolley

	BossSummon

	BossCharge

	BossCharging

	BossStunned
)

type boss struct {
	maxHP    float64
	action   BossAction
	timer    float64
	windup   float64
	cooldown float64
	aim      Vec2
	moves    int
}

type BossStatus struct {
	Health, MaxHealth float64
	Phase             int
	Action            BossAction
	Windup            float64
}

func (g *Game) newBoss() Entity {
	l := g.World.Level
	e := newDemon(centreOf(l.BossAt), Boss)
	e.Z = l.Floor(l.BossAt.X, l.BossAt.Y)
	e.Health = bossHealthBase + bossHealthStep*float64(g.depth/world.ArenaEvery)
	e.alert = true
	g.boss = boss{maxHP: e.Health, cooldown: bossCooldown[0]}
	return e
}

func (g *Game) bossEntity() *Entity {
	for i := range g.Entities {
		if e := &g.Entities[i]; e.Kind == Boss && e.State == Active {
			return e
		}
	}
	return nil
}

func (g *Game) Boss() (BossStatus, bool) {
	e := g.bossEntity()
	if e == nil {
		return BossStatus{}, false
	}
	b := g.boss
	st := BossStatus{Health: e.Health, MaxHealth: b.maxHP, Phase: bossPhase(e.Health, b.maxHP), Action: b.action}
	if b.windup > 0 {
		st.Windup = 1 - max(b.timer, 0)/b.windup
	}
	return st, true
}

func bossPhase(hp, maxHP float64) int {
	switch frac := hp / maxHP; {
	case frac > 2.0/3:
		return 0
	case frac > 1.0/3:
		return 1
	default:
		return 2
	}
}

func (g *Game) updateBoss(e *Entity, i int, pp Vec2, dt float64) {
	b := &g.boss
	e.anim += dt
	e.Frame = int(e.anim * walkFPS)
	g.settleEntity(e, dt)
	if e.hurt > 0 {
		e.hurt -= dt // it shows the hit but never breaks stride for one
	}
	b.timer -= dt
	switch b.action {
	case BossStunned:
		if b.timer <= 0 {
			g.bossRest(e)
		}
	case BossCharging:
		g.bossCharge(e, pp, dt)
	case BossVolley, BossSummon, BossCharge:
		if b.timer <= 0 {
			g.bossStrike(e, i, pp)
		}
	default:
		g.bossStalk(e, i, pp, dt)
	}
}

func (g *Game) bossRest(e *Entity) {
	b := &g.boss
	b.action, b.windup = BossStalk, 0
	b.cooldown = bossCooldown[bossPhase(e.Health, b.maxHP)]
}

func (g *Game) bossStalk(e *Entity, i int, pp Vec2, dt float64) {
	b := &g.boss
	phase := bossPhase(e.Health, b.maxHP)
	d := dist(e.Pos, pp)
	sees := d > 0 && d <= bossSight && losClear(g.World, e.Pos, pp)
	switch {
	case !sees:
		g.stalk(e, i, bossSpeed[phase], dt)
	case d > bossRadius:
		g.advance(e, i, pp, bossSpeed[phase], dt)
	}
	if b.cooldown -= dt; b.cooldown > 0 || !sees {
		return
	}
	moves := bossMoves[phase]
	next := moves[b.moves%len(moves)]
	b.moves++
	if next == BossSummon && g.minions() >= minionCap {
		next = BossVolley
	}
	g.bossWindup(e, next, pp)
}

func (g *Game) bossWindup(e *Entity, action BossAction, pp Vec2) {
	b := &g.boss
	b.action = action
	switch action {
	case BossSummon:
		b.windup = summonWindup
	case BossCharge:
		b.windup = chargeWindup
		// The lane is fixed as the wind-up begins: that is the tell, and the
		// time to step out of it.
		d := dist(e.Pos, pp)
		b.aim = Vec2{X: (pp.X - e.Pos.X) / d, Y: (pp.Y - e.Pos.Y) / d}
	default:
		b.windup = volleyWindup
	}
	b.timer = b.windup
	g.emit(Observation{Kind: ObsBoss, At: cellOf(e.Pos)})
}

func (g *Game) bossStrike(e *Entity, i int, pp Vec2) {
	b := &g.boss
	phase := bossPhase(e.Health, b.maxHP)
	switch b.action {
	case BossCharge:
		b.action, b.timer, b.windup = BossCharging, chargeTime, 0
		return
	case BossSummon:
		g.bossSummon(e, 1+phase)
	default:
		g.bossVolley(e, i, pp, 3+2*phase)
	}
	g.bossRest(e)
}

func (g *Game) bossVolley(e *Entity, i int, pp Vec2, n int) {
	d := dist(e.Pos, pp)
	if d == 0 {
		return
	}
	base := angleOf(Vec2{X: pp.X - e.Pos.X, Y: pp.Y - e.Pos.Y})
	for k := range n {
		a := base + volleySpread*float64(k-n/2)
		target := Vec2{X: e.Pos.X + math.Cos(a)*d, Y: e.Pos.Y + math.Sin(a)*d}
		g.spawnProjectile(i, e.Pos, e.Z+demonEye, target, g.Player.Z+eyeHeight, volleySpeed, volleyDamage)
	}
}

func (g *Game) bossSummon(e *Entity, n int) {
	here := cellOf(e.Pos)
	spawned := 0
	for _, ring := range [2]int{1, 2} {
		for dy := -ring; dy <= ring && spawned < n; dy++ {
			for dx := -ring; dx <= ring && spawned < n; dx++ {
				c := world.Coord{X: here.X + dx, Y: here.Y + dy}
				if chebyInt(c, here) != ring || g.World.Level.At(c.X, c.Y) != world.TileFloor {
					continue
				}
				g.spawnMinion(c, waveKinds[(g.boss.moves+spawned)%(len(waveKinds)-1)])
				spawned++
			}
		}
	}
}

func (g *Game) minions() int {
	n := 0
	for _, e := range g.Entities {
		if e.wave && e.Alive {
			n++
		}
	}
	return n
}

func (g *Game) bossCharge(e *Entity, pp Vec2, dt float64) {
	b := &g.boss
	step := chargeSpeed * dt
	next := g.step(e.Pos, e.Z, b.aim.X*step, b.aim.Y*step)
	moved := dist(next, e.Pos)
	e.Pos = next
	switch {
	case dist(e.Pos, pp) < bossRadius:
		g.faceHurt(e.Pos)
		g.hurtPlayer(chargeDamage)
		g.bossRest(e)
	case moved < step*chargeStall:
		// Ran into something solid: dazed, and soft while it lasts.
		b.action, b.timer, b.windup = BossStunned, stunTime, 0
	case b.timer <= 0:
		g.bossRest(e)
	}
}

func (g *Game) bossDown(e *Entity) {
	for i := range g.Entities {
		if m := &g.Entities[i]; m.wave && m.State == Active {
			m.State, m.Alive, m.anim = Dying, false, 0 // the company falls with its master
		}
	}
	g.Items = append(g.Items,
		ItemState{Kind: world.ItemMega, Pos: e.Pos, Bonus: true},
		ItemState{Kind: world.ItemBackpack, Pos: Vec2{X: e.Pos.X + 0.4, Y: e.Pos.Y}, Bonus: true},
	)
	g.setNotice("The overlord falls. The exit is open")
}

func (g *Game) Sealed() bool { return g.bossEntity() != nil }
