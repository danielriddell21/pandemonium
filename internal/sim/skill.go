package sim

type Skill uint8

const (
	SkillEasy Skill = iota

	SkillNormal

	SkillHard

	SkillNightmare
)

func (s Skill) String() string {
	switch s {
	case SkillEasy:
		return "EASY"
	case SkillHard:
		return "HARD"
	case SkillNightmare:
		return "NIGHTMARE"
	default:
		return "NORMAL"
	}
}

func (s Skill) countScale() float64 {
	switch s {
	case SkillEasy:
		return 0.6
	case SkillHard:
		return 1.4
	case SkillNightmare:
		return 1.8
	default:
		return 1
	}
}

func (s Skill) damageScale() float64 {
	switch s {
	case SkillEasy:
		return 0.5
	case SkillHard:
		return 1.5
	case SkillNightmare:
		return 2
	default:
		return 1
	}
}

func WithSkill(s Skill) Option {
	return func(g *Game) { g.skill = s }
}
