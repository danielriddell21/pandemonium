package sim

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/world"
)

func demonCount(g *Game) int {
	n := 0
	for _, e := range g.Entities {
		if e.Kind != Barrel {
			n++
		}
	}
	return n
}

func TestSkillScalesDemonCount(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 11})
	if err != nil {
		t.Fatal(err)
	}
	easy := demonCount(New(l, WithSkill(SkillEasy)))
	normal := demonCount(New(l, WithSkill(SkillNormal)))
	hard := demonCount(New(l, WithSkill(SkillHard)))
	night := demonCount(New(l, WithSkill(SkillNightmare)))

	if easy > normal || normal > hard || hard > night {
		t.Errorf("counts not monotonic by skill: easy=%d normal=%d hard=%d night=%d", easy, normal, hard, night)
	}
	if night <= easy {
		t.Errorf("nightmare (%d) should field more demons than easy (%d)", night, easy)
	}
}

func TestDefaultSkillIsNormal(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 48, Height: 32, Seed: 11})
	if err != nil {
		t.Fatal(err)
	}
	if def, norm := demonCount(New(l)), demonCount(New(l, WithSkill(SkillNormal))); def != norm {
		t.Errorf("default count %d != normal %d", def, norm)
	}
}

func TestSkillScalesDamage(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 4})
	if err != nil {
		t.Fatal(err)
	}
	lost := func(s Skill) float64 {
		g := New(l, WithSkill(s))
		g.Player.Armor = 0
		before := g.Player.Health
		g.hurtPlayer(20)
		return before - g.Player.Health
	}
	easy, normal, night := lost(SkillEasy), lost(SkillNormal), lost(SkillNightmare)
	if easy >= normal || normal >= night {
		t.Errorf("damage not scaling: easy=%.1f normal=%.1f night=%.1f", easy, normal, night)
	}
}

func TestSkillKeepsGeometry(t *testing.T) {
	l, err := world.Generate(world.Config{Width: 32, Height: 24, Seed: 9})
	if err != nil {
		t.Fatal(err)
	}
	a := New(l, WithSkill(SkillEasy))
	b := New(l, WithSkill(SkillNightmare))
	if a.Player.Pos != b.Player.Pos {
		t.Error("spawn differs by skill")
	}
	if a.World.Level != b.World.Level {
		t.Error("levels should be the same shared geometry")
	}
}
