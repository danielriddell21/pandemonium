package sim

import (
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func TestGunfireRaisesTheAlertAndFistsDoNot(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.Player.Weapon = Fists
	g.fire()
	if g.Alert() != 0 {
		t.Errorf("a punch made noise: alert = %v", g.Alert())
	}
	g.Player.Weapon, g.Player.Shells = Shotgun, 5
	g.fire()
	loud := g.Alert()
	if loud <= 0 {
		t.Fatal("a shotgun blast should raise the alert")
	}
	g.Player.Weapon = Pistol
	g.alarm.meter = 0
	g.fire()
	if g.Alert() >= loud {
		t.Errorf("a pistol (%v) should be quieter than a shotgun (%v)", g.Alert(), loud)
	}
}

func TestAlertSettlesUnderParAndClimbsPastIt(t *testing.T) {
	g := openRoom(t, 20, 12)
	g.alarm.meter = 0.5
	g.tickAlarm(1)
	if g.Alert() >= 0.5 {
		t.Errorf("under par the alert should settle: %v", g.Alert())
	}
	g.elapsed = g.par + 1
	low := g.Alert()
	g.tickAlarm(1)
	if g.Alert() <= low {
		t.Errorf("past par the alert should climb on its own: %v -> %v", low, g.Alert())
	}
	g.alarm.meter = 0
	g.elapsed = 0
	g.tickAlarm(1)
	if g.Alert() != 0 {
		t.Errorf("the alert should not settle below zero: %v", g.Alert())
	}
}

func TestFullAlertCallsAWave(t *testing.T) {
	rec := &recorder{}
	g := openRoom(t, 40, 20, WithObserver(rec))
	g.Player.Pos = Vec2{X: 3.5, Y: 10.5}
	g.trackScent()
	sleeper := demonAt(30.5, 3.5, Melee)
	g.Entities = []Entity{sleeper, newBarrel(Vec2{X: 20.5, Y: 4.5})}
	g.alarm.meter = 1

	g.tickAlarm(1.0 / 60)

	if g.Alarms() != 1 || rec.count(ObsAlarm) != 1 {
		t.Fatalf("alarms = %d, observed %d; want one of each", g.Alarms(), rec.count(ObsAlarm))
	}
	if g.Alert() != 0 {
		t.Errorf("the meter should empty when the alarm sounds: %v", g.Alert())
	}
	if !g.Entities[0].Hunting() {
		t.Error("the alarm should rouse every demon on the level")
	}
	if g.Entities[1].Hunting() {
		t.Error("barrels do not answer alarms")
	}
	wave := g.Entities[2:]
	if len(wave) != g.waveSize() {
		t.Fatalf("wave of %d, want %d", len(wave), g.waveSize())
	}
	for _, e := range wave {
		if !e.wave || !e.Hunting() {
			t.Errorf("reinforcement %+v should be a hunting wave demon", e.Pos)
		}
		if d := g.scentAt(cellOf(e.Pos)); d < waveMinDist {
			t.Errorf("reinforcement arrived %d steps away, want at least %d", d, waveMinDist)
		}
		if e.Kind == Baron {
			t.Error("barons should not answer the first alarm")
		}
	}
	if g.AlarmFlash() <= 0 {
		t.Error("a fresh alarm should flash the screen")
	}
	if g.Notice() == "" {
		t.Error("a fresh alarm should announce itself")
	}
}

func TestAlarmLullStopsTheFightRetriggeringIt(t *testing.T) {
	g := openRoom(t, 40, 20)
	g.trackScent()
	g.alarm.meter = 1
	g.tickAlarm(1.0 / 60)
	g.noise(5)
	g.tickAlarm(1.0 / 60)
	if g.Alarms() != 1 || g.Alert() != 0 {
		t.Errorf("during the lull noise must not count: alarms=%d alert=%v", g.Alarms(), g.Alert())
	}
	g.tickAlarm(alarmLull + 1)
	g.noise(0.2)
	if g.Alert() <= 0 {
		t.Error("after the lull the alert should listen again")
	}
	if g.AlarmFlash() != 0 {
		t.Error("the flash should have faded")
	}
}

func TestWavesEscalate(t *testing.T) {
	g := openRoom(t, 40, 20)
	g.alarm.waves = 1
	first := g.waveSize()
	g.alarm.waves = 3
	if g.waveSize() <= first {
		t.Errorf("third wave (%d) should outnumber the first (%d)", g.waveSize(), first)
	}
	deep := openRoom(t, 40, 20, WithDepth(9))
	deep.alarm.waves = 1
	if deep.waveSize() <= first {
		t.Errorf("a deep first wave (%d) should outnumber a shallow one (%d)", deep.waveSize(), first)
	}
	g.alarm.waves = 99
	if g.waveSize() != waveMax {
		t.Errorf("wave size should cap at %d, got %d", waveMax, g.waveSize())
	}
	if deep.alarmGain() <= g.alarmGain() || openRoom(t, 20, 12, WithDepth(500)).alarmGain() != alarmMaxGain {
		t.Error("alarm sensitivity should grow with depth up to its cap")
	}
}

func TestLaterWavesBringBarons(t *testing.T) {
	g := openRoom(t, 40, 20)
	g.trackScent()
	cells := g.waveCells()
	g.alarm.waves = 3
	g.spawnWave(len(waveKinds), cells)
	barons := 0
	for _, e := range g.Entities {
		if e.Kind == Baron {
			barons++
		}
	}
	if barons == 0 {
		t.Error("a level that keeps ringing should eventually draw a baron")
	}
}

func TestWaveArrivesOutOfSightWhenItCan(t *testing.T) {
	g := openRoom(t, 40, 20)
	for y := 1; y < 15; y++ {
		wall(g, 20, y)
	}
	g.Player.Pos = Vec2{X: 3.5, Y: 3.5}
	g.trackScent()
	cells := g.waveCells()
	if len(cells) == 0 {
		t.Fatal("no cells offered for a wave")
	}
	for _, c := range cells {
		if losClear(g.World, centreOf(c), g.Player.Pos) {
			t.Fatalf("wave cell %v is in plain sight although hidden ones exist", c)
		}
	}
	again := g.waveCells()
	for i := range cells {
		if cells[i] != again[i] {
			t.Fatal("wave cells must be deterministic for a given wave")
		}
	}
}

func TestWaveInAnOpenRoomStillArrives(t *testing.T) {
	g := openRoom(t, 40, 20)
	g.Player.Pos = Vec2{X: 3.5, Y: 10.5}
	g.trackScent()
	if len(g.waveCells()) == 0 {
		t.Error("with nowhere to hide, reinforcements should still arrive at a distance")
	}
}

func TestReinforcementsDropSpoilsOutsideTheTally(t *testing.T) {
	rec := &recorder{}
	g := openRoom(t, 20, 12, WithObserver(rec))
	g.Player.Pos = Vec2{X: 3.5, Y: 5.5}
	g.killsTotal, g.itemsTotal = 0, 0
	g.spawnMinion(world.Coord{X: 3, Y: 5}, Melee)

	g.damageEntity(0, meleeHealth)

	if g.kills != 0 {
		t.Errorf("a reinforcement kill counted toward the level tally: %d", g.kills)
	}
	if rec.count(ObsKill) != 1 {
		t.Error("a reinforcement kill should still be observed")
	}
	if len(g.Items) != 1 || !g.Items[0].Bonus {
		t.Fatalf("want one bonus drop, got %+v", g.Items)
	}
	g.pickupItems()
	if !g.Items[0].Taken || g.items != 0 {
		t.Errorf("spoils should be collectable without counting as level items: taken=%v items=%d", g.Items[0].Taken, g.items)
	}
	if st := g.LevelStats(); st.ItemsPct() != 100 || st.KillsPct() != 100 {
		t.Errorf("tally should be unmoved by spoils: %+v", st)
	}
}

func TestSpoilsCycleThroughTheTable(t *testing.T) {
	g := openRoom(t, 20, 12)
	for range len(spoils) + 1 {
		e := demonAt(5.5, 5.5, Melee)
		g.dropSpoils(&e)
	}
	for i, it := range g.Items {
		if want := spoils[i%len(spoils)]; it.Kind != want {
			t.Errorf("drop %d = %v, want %v", i, it.Kind, want)
		}
	}
}

func TestPanelSilencesTheAlarmOnce(t *testing.T) {
	rec := &recorder{}
	g := openRoom(t, 30, 12, WithObserver(rec))
	l := g.World.Level
	panel := world.Coord{X: 6, Y: 0}
	l.Set(panel.X, panel.Y, world.TileSwitch)
	l.Switches[panel] = world.Switch{Action: world.SwitchAlarm}
	g.Player.Pos, g.Player.Angle = Vec2{X: 6.5, Y: 1.5}, -math.Pi/2
	hunter := demonAt(25.5, 9.5, Melee)
	hunter.alert = true
	g.Entities = []Entity{hunter}
	g.alarm.meter = 0.8

	g.Tick(Input{Interact: true}, 1.0/60)

	if g.Alert() != 0 || !g.AlarmQuiet() || !g.PanelSpent() {
		t.Fatalf("panel should empty the meter and buy quiet: alert=%v quiet=%v spent=%v", g.Alert(), g.AlarmQuiet(), g.PanelSpent())
	}
	if g.Entities[0].Hunting() {
		t.Error("silencing should throw the hunt off the scent")
	}
	if rec.count(ObsSilence) != 1 {
		t.Error("silencing should be observed")
	}
	g.noise(1)
	if g.Alert() != 0 {
		t.Error("noise during the bought quiet should not register")
	}

	g.alarm.quiet = 0
	g.alarm.meter = 0.6
	g.Tick(Input{Interact: true}, 1.0/60)
	if g.Alert() == 0 || rec.count(ObsSilence) != 1 {
		t.Error("a spent panel must not silence a second time")
	}
	if g.Notice() != "The panel is dead" {
		t.Errorf("a spent panel should say so, got %q", g.Notice())
	}
	if g.AlarmQuiet() {
		t.Error("the quiet should have run out")
	}
}

func TestAlarmSleepsOnBossFloors(t *testing.T) {
	g := openRoom(t, 30, 16)
	g.World.Level.Boss = true
	g.noise(5)
	g.elapsed = g.par + 100
	g.tickAlarm(10)
	if g.Alert() != 0 || g.Alarms() != 0 {
		t.Errorf("the overlord's floor has no alarm: alert=%v alarms=%d", g.Alert(), g.Alarms())
	}
}

func TestAlarmCountedInTheTally(t *testing.T) {
	g := openRoom(t, 40, 20)
	g.trackScent()
	g.raiseAlarm()
	g.raiseAlarm()
	if got := g.LevelStats().Alarms; got != 2 {
		t.Errorf("tally alarms = %d, want 2", got)
	}
}

func TestHitRadiusAndAngle(t *testing.T) {
	if hitRadius(Boss) <= hitRadius(Melee) {
		t.Error("the overlord should be a bigger target than its minions")
	}
	if got := angleOf(Vec2{X: 0, Y: 1}); math.Abs(got-math.Pi/2) > 1e-9 {
		t.Errorf("angleOf(+Y) = %v", got)
	}
}

func TestObservationNames(t *testing.T) {
	for k, want := range map[ObservationKind]string{ObsAlarm: "alarm", ObsSilence: "silence", ObsBoss: "boss"} {
		if k.String() != want {
			t.Errorf("%d.String() = %q, want %q", k, k.String(), want)
		}
	}
}
