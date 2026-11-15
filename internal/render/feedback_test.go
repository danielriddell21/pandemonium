package render

import (
	"image/color"
	"math"
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

var testCfg = Config{Width: 256, Height: 160, FOV: 1.152}

func blank(cfg Config) []byte { return make([]byte, cfg.Width*cfg.Height*4) }

func pixel(fb []byte, cfg Config, x, y int) color.RGBA {
	i := (y*cfg.Width + x) * 4
	return color.RGBA{R: fb[i], G: fb[i+1], B: fb[i+2], A: fb[i+3]}
}

func countOf(fb []byte, c color.RGBA) int {
	n := 0
	for i := 0; i+3 < len(fb); i += 4 {
		if fb[i] == c.R && fb[i+1] == c.G && fb[i+2] == c.B {
			n++
		}
	}
	return n
}

func bossGame(t *testing.T) *sim.Game {
	t.Helper()
	l, err := world.Generate(world.Config{Width: 40, Height: 28, Seed: 5, Arena: true})
	if err != nil {
		t.Fatal(err)
	}
	return sim.New(l, sim.WithDepth(world.ArenaEvery-1))
}

func TestPainReddensTheEdgesAndSparesTheCentre(t *testing.T) {
	cfg := testCfg
	fb := blank(cfg)
	viewH := cfg.Height - StatusBarH
	drawPain(fb, cfg, viewH, 1, 0)
	if c := pixel(fb, cfg, cfg.Width/2, viewH/2); c.R != 0 {
		t.Errorf("pain must never cover the middle of the view: %v", c)
	}
	edge := pixel(fb, cfg, 0, viewH/2)
	if edge.R == 0 || edge.R <= edge.G {
		t.Errorf("the screen edge should flush red: %v", edge)
	}
	if below := pixel(fb, cfg, 0, viewH+2); below.R != 0 {
		t.Error("pain must not bleed into the status bar")
	}
	faint := blank(cfg)
	drawPain(faint, cfg, viewH, 0.2, 0)
	if pixel(faint, cfg, 0, viewH/2).R >= edge.R {
		t.Error("a graze should flush less than a mauling")
	}
}

func TestPainLeansTowardWhereTheHitCameFrom(t *testing.T) {
	cfg := testCfg
	fb := blank(cfg)
	viewH := cfg.Height - StatusBarH
	drawPain(fb, cfg, viewH, 1, 1) // struck from the right
	left, right := pixel(fb, cfg, 0, viewH/2), pixel(fb, cfg, cfg.Width-1, viewH/2)
	if right.R <= left.R {
		t.Errorf("a hit from the right should flush the right edge harder: left %d right %d", left.R, right.R)
	}
}

func TestHitMarkerIsADiagonalCrossClearOfTheCentre(t *testing.T) {
	cfg := testCfg
	fb := blank(cfg)
	viewH := cfg.Height - StatusBarH
	drawHitMarker(fb, cfg, viewH)
	cx, cy := cfg.Width/2, viewH/2
	if got := countOf(fb, hitWhite); got != 16 {
		t.Errorf("hit marker drew %d pixels, want 4 arms of 4", got)
	}
	if pixel(fb, cfg, cx, cy).R != 0 {
		t.Error("the marker must leave the aim point clear")
	}
	if pixel(fb, cfg, cx+4, cy+4) != hitWhite || pixel(fb, cfg, cx-4, cy+4) != hitWhite {
		t.Error("the marker arms should sit on the diagonals")
	}
}

func TestEdgeBarsFrameTheViewOnly(t *testing.T) {
	cfg := testCfg
	fb := blank(cfg)
	viewH := cfg.Height - StatusBarH
	drawEdgeBars(fb, cfg, viewH, alertHot, 1)
	if pixel(fb, cfg, 10, 0) != alertHot || pixel(fb, cfg, 10, viewH-1) != alertHot {
		t.Error("alarm bars should run along the top and bottom of the view")
	}
	if pixel(fb, cfg, 10, viewH/2).R != 0 || pixel(fb, cfg, 10, viewH+1).R != 0 {
		t.Error("alarm bars must not cover the view or the status bar")
	}
}

func TestBlendPixel(t *testing.T) {
	cfg := Config{Width: 2, Height: 1}
	fb := []byte{100, 100, 100, 255, 0, 0, 0, 0}
	blendPixel(fb, 2, 0, 0, color.RGBA{R: 200, G: 0, B: 100}, 0.5)
	if got := pixel(fb, cfg, 0, 0); got != (color.RGBA{R: 150, G: 50, B: 100, A: 255}) {
		t.Errorf("half blend = %v", got)
	}
	blendPixel(fb, 2, 1, 0, color.RGBA{R: 200}, 7) // over-strength clamps to opaque
	if got := pixel(fb, cfg, 1, 0); got.R != 200 || got.A != 255 {
		t.Errorf("over-strength blend = %v", got)
	}
	before := append([]byte(nil), fb...)
	blendPixel(fb, 2, 0, 0, color.RGBA{R: 9}, 0)
	blendPixel(fb, 2, 5, 5, color.RGBA{R: 9}, 1)
	for i := range fb {
		if fb[i] != before[i] {
			t.Fatal("a zero-strength or out-of-bounds blend must change nothing")
		}
	}
}

func TestMeterFillsInProportion(t *testing.T) {
	cfg := testCfg
	fill := color.RGBA{R: 1, G: 2, B: 3, A: 255}
	half := blank(cfg)
	x0, bw := drawMeter(half, cfg, "X", 0.5, fill, palette.hudText)
	if got, want := countOf(half, fill), bw/2*threatBarH; got != want {
		t.Errorf("half meter filled %d pixels, want %d", got, want)
	}
	if x0 != (cfg.Width-bw)/2 {
		t.Errorf("meter is not centred: x0=%d bw=%d", x0, bw)
	}
	over := blank(cfg)
	drawMeter(over, cfg, "X", 9, fill, palette.hudText)
	if got := countOf(over, fill); got != bw*threatBarH {
		t.Errorf("an over-full meter should clamp to its trough: %d", got)
	}
	if countOf(blank(cfg), barTrough) != 0 || countOf(half, barTrough) == 0 {
		t.Error("the empty part of the meter should show its trough")
	}
}

func TestThreatBarStaysOutOfTheWayUntilThereIsAThreat(t *testing.T) {
	cfg := testCfg
	g := smokeGame(t)
	g.Entities = nil
	calm := blank(cfg)
	drawThreatBar(calm, cfg, g)
	if countColored(calm, cfg) != 0 {
		t.Fatal("a quiet level should show no alert meter")
	}

	g.Player.Weapon, g.Player.Shells = sim.Shotgun, 50
	for g.Alert() < 0.3 {
		g.Tick(sim.Input{Attack: true}, 1.0/60)
	}
	warm := blank(cfg)
	drawThreatBar(warm, cfg, g)
	if countOf(warm, alertCalm) == 0 {
		t.Error("gunfire should bring up the alert meter")
	}

	for g.Alert() < hotAlert {
		g.Tick(sim.Input{Attack: true}, 1.0/60)
	}
	hot := blank(cfg)
	drawThreatBar(hot, cfg, g)
	if countOf(hot, alertHot) == 0 || countOf(hot, alertCalm) != 0 {
		t.Error("a nearly full meter should turn hot")
	}
}

func TestAlarmFlashesAndTheMeterLingersAfterIt(t *testing.T) {
	cfg := testCfg
	g := smokeGame(t)
	g.Entities = nil
	g.Player.Weapon, g.Player.Shells = sim.Shotgun, 50
	g.Player.InvulnTTL = 1e6
	for g.Alarms() == 0 {
		g.Tick(sim.Input{Attack: true}, 1.0/60)
	}
	if g.AlarmFlash() <= 0 {
		t.Fatal("a fresh alarm should flash")
	}
	// The bars strobe: over a dozen ticks at least one frame shows them.
	shown := false
	for range 12 {
		fb := blank(cfg)
		drawFeedback(fb, cfg, g)
		shown = shown || pixel(fb, cfg, 10, 0).R > 0
		g.Tick(sim.Input{}, 1.0/60)
	}
	if !shown {
		t.Error("the alarm should strobe the edges of the view")
	}
	fb := blank(cfg)
	drawThreatBar(fb, cfg, g)
	if countColored(fb, cfg) == 0 {
		t.Error("once an alarm has sounded the meter should stay up, even empty")
	}
}

func TestFeedbackShowsPainAndHits(t *testing.T) {
	cfg := testCfg
	g := smokeGame(t)
	g.Entities = []sim.Entity{{Pos: g.Player.Pos, Z: g.Player.Z, Kind: sim.Melee, State: sim.Active, Health: 500, Alive: true}}
	rest := blank(cfg)
	drawFeedback(rest, cfg, g)
	if countColored(rest, cfg) != 0 {
		t.Fatal("at rest there is nothing to show")
	}
	for range 30 {
		g.Tick(sim.Input{}, 1.0/60) // the demon on top of the player claws them
	}
	hurt := blank(cfg)
	drawFeedback(hurt, cfg, g)
	if pixel(hurt, cfg, 0, 40).R == 0 {
		t.Error("taking damage should flush the edges")
	}

	g.Entities[0].Pos = sim.Vec2{X: g.Player.Pos.X + math.Cos(g.Player.Angle), Y: g.Player.Pos.Y + math.Sin(g.Player.Angle)}
	g.Player.Weapon = sim.Fists
	for !g.HitMarker() {
		g.Tick(sim.Input{Attack: true}, 1.0/60)
	}
	hit := blank(cfg)
	drawFeedback(hit, cfg, g)
	if countOf(hit, hitWhite) == 0 {
		t.Error("a landed blow should show the hit marker")
	}
}

func TestBossBarReplacesTheAlertMeter(t *testing.T) {
	cfg := testCfg
	g := bossGame(t)
	fb := blank(cfg)
	drawThreatBar(fb, cfg, g)
	st, ok := g.Boss()
	if !ok {
		t.Fatal("boss floor without a boss")
	}
	full := color.RGBA{R: 214, G: 120, B: 40, A: 255}
	bw := min(cfg.Width/2, 220)
	if got := countOf(fb, full); got < (bw-2)*threatBarH {
		t.Errorf("a fresh boss should show a full first-phase bar: %d px", got)
	}
	// Phase notches are cut out of the fill.
	x0 := (cfg.Width - bw) / 2
	if pixel(fb, cfg, x0+bw/3, threatTop+1) != barTrough {
		t.Error("the bar should be notched where the next phase begins")
	}

	wounded, stunned := blank(cfg), blank(cfg)
	st.Health, st.Phase = st.MaxHealth*0.2, 2
	drawBossBar(wounded, cfg, st, 0)
	if countOf(wounded, full) != 0 {
		t.Error("a last-phase bar should change colour")
	}
	st.Action = sim.BossStunned
	drawBossBar(stunned, cfg, st, 0)
	if countOf(stunned, tintStun) == 0 {
		t.Error("a stunned boss should flash its bar: that is the opening")
	}
}

func TestBossBarNamesEachTell(t *testing.T) {
	cfg := testCfg
	base := sim.BossStatus{Health: 500, MaxHealth: 1000, Phase: 1}
	idle := blank(cfg)
	drawBossBar(idle, cfg, base, 1)
	for _, a := range []sim.BossAction{sim.BossVolley, sim.BossSummon, sim.BossCharge, sim.BossCharging, sim.BossStunned} {
		st := base
		st.Action = a
		fb := blank(cfg)
		drawBossBar(fb, cfg, st, 1)
		same := true
		for i := range fb {
			if fb[i] != idle[i] {
				same = false
				break
			}
		}
		if same {
			t.Errorf("action %d reads the same as stalking", a)
		}
	}
}

func TestBossTintTracksTheWindup(t *testing.T) {
	if _, mix := bossTint(sim.BossStatus{Action: sim.BossStalk}, 0); mix != 0 {
		t.Error("a stalking boss is untinted")
	}
	_, early := bossTint(sim.BossStatus{Action: sim.BossVolley, Windup: 0.1}, 1)
	c, late := bossTint(sim.BossStatus{Action: sim.BossVolley, Windup: 0.9}, 1)
	if late <= early || c != tintWindup {
		t.Errorf("the tell should brighten through the wind-up: %v -> %v (%v)", early, late, c)
	}
	_, on := bossTint(sim.BossStatus{Action: sim.BossVolley, Windup: 0.9}, 0)
	_, off := bossTint(sim.BossStatus{Action: sim.BossVolley, Windup: 0.9}, 4)
	if on == off {
		t.Error("the last of the wind-up should strobe")
	}
	if c, _ := bossTint(sim.BossStatus{Action: sim.BossCharge, Windup: 0.5}, 1); c != tintCharge {
		t.Error("a charge tell should be red, apart from the others")
	}
	if c, mix := bossTint(sim.BossStatus{Action: sim.BossCharging}, 1); c != tintCharge || mix == 0 {
		t.Error("a charging boss stays red")
	}
	if c, mix := bossTint(sim.BossStatus{Action: sim.BossStunned}, 1); c != tintStun || mix == 0 {
		t.Error("a stunned boss reads blue")
	}
}

func TestDemonBillboards(t *testing.T) {
	tx := defaultTextures()
	g := bossGame(t)
	var boss sim.Entity
	for _, e := range g.Entities {
		if e.Kind == sim.Boss {
			boss = e
		}
	}
	b := demonBillboard(g, tx, boss)
	if b.scale != bossScale || b.mix != 0 {
		t.Errorf("a resting boss should be large and untinted: %+v", b)
	}
	if b.tex != tx.demon[len(tx.demon)-1].walk[0] {
		t.Error("the boss should wear the overlord art")
	}
	imp := sim.Entity{Kind: sim.Ranged, Sprite: 1, State: sim.Active, Alive: true}
	if got := demonBillboard(g, tx, imp); got.scale != spriteScale || got.mix != 0 {
		t.Errorf("an ordinary demon: %+v", got)
	}
}

func TestFlinchingDemonFlashes(t *testing.T) {
	cfg := testCfg
	g := smokeGame(t)
	ahead := sim.Vec2{X: g.Player.Pos.X + 1.2*math.Cos(g.Player.Angle), Y: g.Player.Pos.Y + 1.2*math.Sin(g.Player.Angle)}
	g.Entities = []sim.Entity{{Pos: ahead, Z: g.Player.Z, Kind: sim.Baron, Sprite: 4, State: sim.Active, Health: 5000, Alive: true}}
	r := NewRenderer(cfg)
	r.SetHUD(false)
	calm := append([]byte(nil), r.Frame(g)...)
	g.Player.Weapon = sim.Fists
	for !g.Entities[0].Flinching() {
		g.Tick(sim.Input{Attack: true}, 1.0/60)
	}
	if b := demonBillboard(g, defaultTextures(), g.Entities[0]); b.tint != tintFlinch || b.mix != flinchMix {
		t.Fatalf("a struck demon should flash: %+v", b)
	}
	if brightnessSum(r.Frame(g)) <= brightnessSum(calm) {
		t.Error("the flash should brighten the demon on screen")
	}
}

func TestPanelTexturesAndSwitchFaces(t *testing.T) {
	tx := defaultTextures()
	lamp := color.RGBA{R: 235, G: 60, B: 40, A: 255}
	live, dead := 0, 0
	for _, c := range tx.panel.pix {
		if c == lamp {
			live++
		}
	}
	for _, c := range tx.panelDead.pix {
		if c == lamp {
			dead++
		}
	}
	if live == 0 || dead != 0 {
		t.Errorf("a live panel shows a red lamp (%d px), a spent one does not (%d px)", live, dead)
	}

	g := smokeGame(t)
	l := g.World.Level
	p := &columnPainter{g: g, tx: tx}
	panel, ok := l.AlarmPanel()
	if !ok {
		t.Fatal("seed 7 should carry an alarm panel")
	}
	if p.switchTexture(panel) != tx.panel {
		t.Error("the alarm panel should wear its own face")
	}
	for c, sw := range l.Switches {
		if sw.Action == world.SwitchExit && p.switchTexture(c) != tx.switchTex {
			t.Error("the exit switch should keep its lever")
		}
	}
	// Walk up to the panel and use it: the lamp goes out.
	for _, n := range [4]world.Coord{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}} {
		f := world.Coord{X: panel.X + n.X, Y: panel.Y + n.Y}
		if l.At(f.X, f.Y) != world.TileFloor {
			continue
		}
		g.Entities = nil
		g.Player.Pos = sim.Vec2{X: float64(f.X) + 0.5, Y: float64(f.Y) + 0.5}
		g.Player.Z = l.Floor(f.X, f.Y)
		g.Player.Angle = math.Atan2(float64(-n.Y), float64(-n.X))
	}
	prompt := blank(testCfg)
	drawInteractPrompt(prompt, testCfg, g)
	if countColored(prompt, testCfg) == 0 {
		t.Error("facing the panel should prompt to silence the alarm")
	}
	g.Tick(sim.Input{Interact: true}, 1.0/60)
	if p.switchTexture(panel) != tx.panelDead {
		t.Error("a used panel should go dark")
	}
	spent := blank(testCfg)
	g.Tick(sim.Input{}, 3) // let the notice lapse
	drawInteractPrompt(spent, testCfg, g)
	if countColored(spent, testCfg) != 0 {
		t.Error("a spent panel should not prompt")
	}
}

func TestOverlordArtIsDistinct(t *testing.T) {
	tx := defaultTextures()
	boss := tx.demon[len(tx.demon)-1]
	horn := color.RGBA{R: 232, G: 214, B: 170, A: 255}
	horns := 0
	for _, c := range boss.walk[0].pix {
		if c == horn {
			horns++
		}
	}
	if horns == 0 {
		t.Error("the overlord should have horns")
	}
	if boss.walk[0].at(2, 2).A != 0 {
		t.Error("the overlord sprite needs a transparent background")
	}
	if len(boss.walk) != 2 || len(boss.dead) != 3 {
		t.Errorf("overlord animation frames: %d walk, %d dead", len(boss.walk), len(boss.dead))
	}
}

func TestBevelLightsOneEdgeAndShadesTheOther(t *testing.T) {
	tex := newTexture(8, 8)
	base := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	fillRect(tex, 2, 2, 6, 6, base)
	bevel(tex, 2, 2, 6, 6, base)
	if tex.at(2, 3).R <= base.R || tex.at(5, 3).R >= base.R || tex.at(3, 3) != base {
		t.Errorf("bevel edges: left %v right %v middle %v", tex.at(2, 3), tex.at(5, 3), tex.at(3, 3))
	}
}

func TestStatusBarColumnsNeverCollide(t *testing.T) {
	for _, w := range []int{256, 320, 640, 1280} {
		face := StatusBarH - 6
		keys := w - w/32 - len(keyColors)*keyPitch
		c := barColumns(w, face, keys)
		healthEnd := c.health + len("HEALTH")*glyphWidth
		armorEnd := c.armor + len("ARMOR")*glyphWidth
		ammoEnd := c.ammo + len("AMMO")*glyphWidth
		if healthEnd > c.armor {
			t.Errorf("width %d: HEALTH (to %d) runs into ARMOR (%d)", w, healthEnd, c.armor)
		}
		if armorEnd > w/2-face/2 {
			t.Errorf("width %d: ARMOR (to %d) runs under the face (%d)", w, armorEnd, w/2-face/2)
		}
		if c.ammo < w/2+face/2 {
			t.Errorf("width %d: AMMO (%d) starts under the face (%d)", w, c.ammo, w/2+face/2)
		}
		if ammoEnd > keys {
			t.Errorf("width %d: AMMO (to %d) runs under the keycards (%d)", w, ammoEnd, keys)
		}
	}
}

func TestPlateDimsWhatIsBehindText(t *testing.T) {
	cfg := testCfg
	fb := blank(cfg)
	for i := range fb {
		fb[i] = 200
	}
	drawPlate(fb, cfg, 100, 80, 40)
	if got := pixel(fb, cfg, 110, 76); got.R >= 200 || got.R == 0 {
		t.Errorf("behind the text should be dimmed, not blacked out: %v", got)
	}
	if pixel(fb, cfg, 10, 76).R != 200 || pixel(fb, cfg, 110, 20).R != 200 {
		t.Error("the plate should stay local to the text")
	}
	drawPlate(fb, cfg, -50, 2, 400) // partly off-screen must not panic
}

func TestWeaponRestsWhenStandingAndSwingsWithTheStride(t *testing.T) {
	cfg := Config{Width: 200, Height: 120, FOV: 1.152}
	tx := defaultTextures()
	draw := func(stride float64, firing bool) []byte {
		fb := blank(cfg)
		drawViewmodel(fb, cfg, tx.weapon[1], tx.flash, firing, stride, cfg.Height)
		return fb
	}
	same := func(a, b []byte) bool {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	if !same(draw(3.3, false), draw(3.3, false)) {
		t.Error("a player standing still must hold a steady weapon")
	}
	if same(draw(0, false), draw(0.5, false)) {
		t.Error("walking should swing the weapon")
	}
	if same(draw(0, false), draw(0, true)) {
		t.Error("firing should kick the weapon")
	}
}

func TestTallyShowsAlarms(t *testing.T) {
	cfg := testCfg
	quiet, loud := blank(cfg), blank(cfg)
	drawIntermission(quiet, cfg, sim.LevelStats{})
	drawIntermission(loud, cfg, sim.LevelStats{Alarms: 3})
	if countOf(quiet, alertHot) != 0 {
		t.Error("a quiet floor should not be marked hot")
	}
	if countOf(loud, alertHot) == 0 {
		t.Error("a floor that rang should say so in the tally")
	}
	if countOf(quiet, barBG) == 0 {
		t.Error("the tally should sit on its panel")
	}
}

func TestAutomapMarksThePanelAndTheExitSwitch(t *testing.T) {
	l := smokeGame(t).World.Level
	panel, _ := l.AlarmPanel()
	if c, ok := tileColor(l, panel); !ok || c != mapAlarm {
		t.Errorf("alarm panel on the map = %v, %v", c, ok)
	}
	for at, sw := range l.Switches {
		if sw.Action != world.SwitchExit {
			continue
		}
		if c, ok := tileColor(l, at); !ok || c != mapExit {
			t.Errorf("exit switch on the map = %v, %v", c, ok)
		}
	}
}
