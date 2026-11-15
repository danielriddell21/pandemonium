package gui

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

func TestSessionCarriesTheKitDownAndDeepens(t *testing.T) {
	s := newSession(7, 32, 24, telemetry.NewBus(), nil)
	first := s.start()
	if first.Depth() != 0 || first.Loadout().Bullets != 50 {
		t.Fatalf("first level: depth %d, loadout %+v", first.Depth(), first.Loadout())
	}
	first.Player.Bullets, first.Player.Rockets, first.Player.Armor = 3, 11, 40
	first.Player.Health = 20
	first.Player.Weapon = sim.Shotgun

	second := s.next()
	if second == first {
		t.Fatal("next should build a new level")
	}
	p := second.Player
	if p.Bullets != 3 || p.Rockets != 11 || p.Armor != 40 || p.Weapon != sim.Shotgun {
		t.Errorf("kit did not follow the player down: %+v", p)
	}
	if p.Health != sim.MaxHealth {
		t.Errorf("wounds should be patched between floors: %v", p.Health)
	}
	if second.Depth() != 1 {
		t.Errorf("second level depth = %d, want 1", second.Depth())
	}
}

func TestSessionFifthLevelIsTheBossFloor(t *testing.T) {
	s := newSession(7, 32, 24, telemetry.NewBus(), nil)
	g := s.start()
	for range 4 {
		if g.World.Level.Boss {
			t.Fatal("a boss floor came early")
		}
		g = s.next()
	}
	if !g.World.Level.Boss || !g.Sealed() {
		t.Error("the fifth level should be a sealed boss floor")
	}
}

func TestSessionSkillReachesTheLevel(t *testing.T) {
	s := newSession(7, 32, 24, telemetry.NewBus(), nil)
	s.setSkill(int(sim.SkillNightmare))
	hard := len(s.start().Entities)
	e := newSession(7, 32, 24, telemetry.NewBus(), nil)
	e.setSkill(99) // out of range falls back to normal
	if normal := len(e.start().Entities); hard <= normal {
		t.Errorf("nightmare should field more demons than normal: %d vs %d", hard, normal)
	}
}

func TestAnyHeld(t *testing.T) {
	if anyHeld() || anyHeld(false, false) {
		t.Error("nothing held should not fire")
	}
	if !anyHeld(false, true, false) {
		t.Error("any one trigger held should fire")
	}
}
