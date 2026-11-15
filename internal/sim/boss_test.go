package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestBossFloorSpawnsOneOverlordAndAThinEscort(t *testing.T) {
	l := roomLevel(40, 20)
	l.Boss, l.BossAt = true, world.Coord{X: 30, Y: 10}
	g := New(l)
	bosses, escort := 0, 0
	for _, e := range g.Entities {
		switch e.Kind {
		case Boss:
			bosses++
			if cellOf(e.Pos) != l.BossAt {
				t.Errorf("boss spawned at %v, want %v", cellOf(e.Pos), l.BossAt)
			}
		case Barrel:
		default:
			escort++
			if cellOf(e.Pos) == l.BossAt {
				t.Error("an escort demon spawned on the boss")
			}
		}
	}
	if bosses != 1 || escort != bossEscort {
		t.Errorf("bosses=%d escort=%d, want 1 and %d", bosses, escort, bossEscort)
	}
	if st, ok := g.Boss(); !ok || st.Health != bossHealthBase || st.MaxHealth != bossHealthBase || st.Phase != 0 {
		t.Errorf("fresh boss status = %+v ok=%v", st, ok)
	}
	if hard := New(l, WithSkill(SkillNightmare)); entityCount(l, hard.skill, 0) <= bossEscort {
		t.Error("a harder skill should thicken the escort")
	}
}

func TestBossGrowsTougherEachTier(t *testing.T) {
	_, first := bossRoom(t, WithDepth(world.ArenaEvery-1))
	_, second := bossRoom(t, WithDepth(2*world.ArenaEvery-1))
	if second.Health <= first.Health {
		t.Errorf("second overlord (%v hp) should outlast the first (%v hp)", second.Health, first.Health)
	}
}

func TestNoBossOnAnOrdinaryFloor(t *testing.T) {
	g := openRoom(t, 20, 12)
	if _, ok := g.Boss(); ok || g.Sealed() {
		t.Error("an ordinary floor has no boss and no seal")
	}
}

func TestBossPhases(t *testing.T) {
	for _, c := range []struct {
		hp   float64
		want int
	}{{900, 0}, {601, 0}, {600, 1}, {301, 1}, {300, 2}, {1, 2}} {
		if got := bossPhase(c.hp, 900); got != c.want {
			t.Errorf("bossPhase(%v/900) = %d, want %d", c.hp, got, c.want)
		}
	}
}

func TestExitIsSealedWhileTheBossLives(t *testing.T) {
	g, b := bossRoom(t)
	g.pressSwitch(world.Switch{Action: world.SwitchExit})
	if g.LevelComplete() || !g.Sealed() {
		t.Fatal("the exit switch must do nothing while the overlord lives")
	}
	if g.Notice() == "" {
		t.Error("a sealed exit should say why")
	}
	b.Health = 1
	g.damageEntity(0, 5)
	if g.Sealed() {
		t.Fatal("killing the overlord should break the seal")
	}
	g.World.Level.Switches[world.Coord{X: 1, Y: 0}] = world.Switch{Action: world.SwitchExit}
	g.pressSwitch(world.Switch{Action: world.SwitchExit})
	if !g.LevelComplete() {
		t.Error("the exit should open once the overlord is dead")
	}
}

func TestBossTelegraphsThenVolleys(t *testing.T) {
	rec := &recorder{}
	g, _ := bossRoom(t, WithObserver(rec))
	g.Player.Pos = Vec2{X: 22.5, Y: 10.5}
	g.Player.InvulnTTL = 1000

	sawWindup := false
	for range 60 * 6 {
		g.Tick(Input{}, 1.0/60)
		st, _ := g.Boss()
		if st.Action == BossVolley {
			if len(g.Projectiles) != 0 && !sawWindup {
				t.Fatal("fireballs flew before any tell")
			}
			sawWindup = true
			if st.Windup < 0 || st.Windup > 1 {
				t.Fatalf("windup progress %v out of range", st.Windup)
			}
		}
		if sawWindup && st.Action == BossStalk {
			break
		}
	}
	if !sawWindup {
		t.Fatal("the boss never wound up an attack")
	}
	if len(g.Projectiles) != 3 {
		t.Fatalf("a first-phase volley is 3 fireballs, got %d", len(g.Projectiles))
	}
	if rec.count(ObsBoss) == 0 {
		t.Error("a wind-up should be observed, so it can be heard")
	}
	// The spread fans out: no two fireballs share a heading.
	a, b := angleOf(g.Projectiles[0].Vel), angleOf(g.Projectiles[1].Vel)
	if math.Abs(a-b) < volleySpread/2 {
		t.Errorf("volley did not spread: headings %v and %v", a, b)
	}
}

func TestBossVolleyWidensWithPhase(t *testing.T) {
	g, b := bossRoom(t)
	g.Player.Pos = Vec2{X: 22.5, Y: 10.5}
	b.Health = g.boss.maxHP * 0.1
	g.boss.action, g.boss.timer = BossVolley, 0
	g.bossStrike(b, 0, g.Player.Pos)
	if len(g.Projectiles) != 7 {
		t.Errorf("a last-phase volley is 7 fireballs, got %d", len(g.Projectiles))
	}
	if g.boss.action != BossStalk || g.boss.cooldown != bossCooldown[2] {
		t.Errorf("after striking the boss should rest on its phase cooldown: %+v", g.boss)
	}
}

func TestBossSummonsACappedCompany(t *testing.T) {
	g, b := bossRoom(t)
	b.Health = g.boss.maxHP * 0.5
	g.boss.action, g.boss.timer = BossSummon, 0
	g.bossStrike(&g.Entities[0], 0, g.Player.Pos)
	if got := g.minions(); got != 2 {
		t.Fatalf("a second-phase summon calls 2, got %d", got)
	}
	for _, e := range g.Entities[1:] {
		if !e.wave || !e.Hunting() || e.Kind == Baron || e.Kind == Boss {
			t.Errorf("summoned %+v should be a hunting rank-and-file minion", e.Kind)
		}
		if chebyInt(cellOf(e.Pos), cellOf(g.Entities[0].Pos)) > 2 {
			t.Errorf("minion summoned far from the boss at %v", e.Pos)
		}
	}

	// At the cap, the summon turn is spent on a volley instead.
	for g.minions() < minionCap {
		g.spawnMinion(world.Coord{X: 5, Y: 5}, Melee)
	}
	g.Player.Pos = Vec2{X: 24.5, Y: 10.5}
	g.boss.action, g.boss.cooldown, g.boss.moves = BossStalk, 0, 1
	g.bossStalk(&g.Entities[0], 0, g.Player.Pos, 1.0/60)
	if g.boss.action != BossVolley {
		t.Errorf("with the company full the boss should volley, not summon: action %d", g.boss.action)
	}
}

func TestBossChargeLaneIsFixedAtTheTell(t *testing.T) {
	g, b := bossRoom(t)
	g.Player.Pos = Vec2{X: 22.5, Y: 10.5}
	g.bossWindup(b, BossCharge, g.Player.Pos)
	aim := g.boss.aim
	if math.Abs(aim.X+1) > 1e-9 || math.Abs(aim.Y) > 1e-9 {
		t.Fatalf("charge lane %v should point at the player", aim)
	}

	// The player steps aside during the wind-up; the lane must not follow.
	g.Player.Pos = Vec2{X: 22.5, Y: 15.5}
	g.Player.InvulnTTL = 1000
	start := b.Pos
	for range 60 * 3 {
		g.Tick(Input{}, 1.0/60)
		if st, _ := g.Boss(); st.Action == BossStalk || st.Action == BossStunned {
			break
		}
	}
	if g.boss.aim != aim {
		t.Error("the charge lane moved after the tell")
	}
	if math.Abs(b.Pos.Y-start.Y) > 0.01 || b.Pos.X >= start.X-3 {
		t.Errorf("the charge should run straight down its lane: %v -> %v", start, b.Pos)
	}
	if g.Player.Health != MaxHealth {
		t.Error("a sidestepped charge must not hurt")
	}
}

func TestBossChargeHitsWhoStaysInTheLane(t *testing.T) {
	g, b := bossRoom(t)
	g.Player.Pos = Vec2{X: 22.5, Y: 10.5}
	g.bossWindup(b, BossCharge, g.Player.Pos)
	g.boss.action, g.boss.timer = BossCharging, chargeTime
	before := g.Player.Health
	for range 120 {
		g.bossCharge(b, g.Player.Pos, 1.0/60)
		if g.boss.action != BossCharging {
			break
		}
	}
	if got := before - g.Player.Health; got != chargeDamage {
		t.Errorf("a charge that connects should deal %v, dealt %v", chargeDamage, got)
	}
	if g.boss.action != BossStalk {
		t.Errorf("after connecting the boss should go back to stalking: %d", g.boss.action)
	}
}

func TestBossChargedIntoAWallIsStunnedAndSoft(t *testing.T) {
	g, b := bossRoom(t)
	g.Player.Pos = Vec2{X: 5.5, Y: 3.5}
	g.boss.aim = Vec2{X: 1, Y: 0} // straight at the east wall
	g.boss.action, g.boss.timer = BossCharging, chargeTime
	for range 120 {
		g.bossCharge(b, g.Player.Pos, 1.0/60)
		if g.boss.action != BossCharging {
			break
		}
	}
	if st, _ := g.Boss(); st.Action != BossStunned {
		t.Fatalf("a charge into a wall should stun, action = %d", st.Action)
	}
	hp := b.Health
	g.damageEntity(0, 10)
	if got := hp - b.Health; got != 10*stunWeak {
		t.Errorf("a stunned boss should take %vx damage: 10 dealt %v", stunWeak, got)
	}
	pos := b.Pos
	for range int(stunTime*60) + 2 {
		g.updateBoss(b, 0, g.Player.Pos, 1.0/60)
	}
	if g.boss.action == BossStunned {
		t.Error("the stun should wear off")
	}
	if b.Pos != pos {
		t.Error("a stunned boss must not move")
	}
}

func TestBossChargeTimesOut(t *testing.T) {
	g, b := bossRoom(t)
	g.Player.Pos = Vec2{X: 5.5, Y: 3.5}
	g.boss.aim = Vec2{X: -1, Y: 0}
	g.boss.action, g.boss.timer = BossCharging, -0.1
	g.bossCharge(b, g.Player.Pos, 1.0/60)
	if g.boss.action != BossStalk {
		t.Errorf("a charge that runs out of time should end: %d", g.boss.action)
	}
}

func TestBossDeathFellsTheCompanyAndDropsTheHaul(t *testing.T) {
	g, _ := bossRoom(t)
	g.spawnMinion(world.Coord{X: 10, Y: 10}, Melee)
	g.Entities[0].Health = 1
	g.killsTotal = 1

	g.damageEntity(0, 5)

	if g.Entities[1].Alive {
		t.Error("the company should fall with its master")
	}
	kinds := map[world.ItemKind]bool{}
	for _, it := range g.Items {
		if !it.Bonus {
			t.Errorf("boss loot %v should sit outside the tally", it.Kind)
		}
		kinds[it.Kind] = true
	}
	if !kinds[world.ItemMega] || !kinds[world.ItemBackpack] {
		t.Errorf("boss should drop a megasphere and a backpack, got %v", g.Items)
	}
	if g.kills != 1 {
		t.Errorf("the overlord counts as a kill: %d", g.kills)
	}
}

func TestBossIgnoresFlinchingAndTouchesFromFurther(t *testing.T) {
	g, b := bossRoom(t)
	g.Player.Pos = Vec2{X: b.Pos.X - 0.8, Y: b.Pos.Y}
	g.damageEntity(0, 1)
	if !b.Flinching() {
		t.Fatal("a hit should show on the boss")
	}
	before := g.Player.Health
	g.applyContactDamage(1.0 / 60)
	if g.Player.Health >= before {
		t.Error("the overlord's bulk should reach a player a minion could not")
	}
	if touchRange(Boss) <= touchRange(Melee) {
		t.Error("boss reach should exceed a minion's")
	}
	pos := b.Pos
	g.Player.Pos = Vec2{X: 20.5, Y: 10.5}
	g.updateBoss(b, 0, g.Player.Pos, 1.0/60)
	if b.Pos == pos {
		t.Error("a hit must not stop the boss in its tracks")
	}
}

func TestBossHuntsWhatItCannotSee(t *testing.T) {
	g, b := bossRoom(t)
	for y := 1; y < 15; y++ {
		wall(g, 20, y)
	}
	g.Player.Pos = Vec2{X: 5.5, Y: 3.5}
	g.trackScent()
	start := b.Pos
	for range 120 {
		g.updateBoss(b, 0, g.Player.Pos, 1.0/60)
	}
	if b.Pos == start {
		t.Error("the overlord should come looking round the wall")
	}
	if g.boss.action != BossStalk {
		t.Error("it should not attack what it cannot see")
	}
}

func TestBossVolleyOnTopOfThePlayerIsANoOp(t *testing.T) {
	g, b := bossRoom(t)
	g.bossVolley(b, 0, b.Pos, 3)
	if len(g.Projectiles) != 0 {
		t.Error("a volley with no direction should fire nothing")
	}
}
