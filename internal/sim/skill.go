package sim

// Skill is the difficulty level. It scales how many demons spawn and how hard
// their attacks land, without touching the generated geometry — the same seed
// yields the same map at every skill, just a harder or easier fight.
type Skill uint8

const (
	// SkillEasy thins the demons out and softens their hits.
	SkillEasy Skill = iota
	// SkillNormal is the baseline the game is tuned around.
	SkillNormal
	// SkillHard packs in more demons that hit harder.
	SkillHard
	// SkillNightmare is the punishing extreme.
	SkillNightmare
)

// String returns the display name.
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

// countScale multiplies the demon population for this skill.
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

// damageScale multiplies the damage demons deal to the player.
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

// WithSkill sets the difficulty for the simulation. The default is SkillNormal.
func WithSkill(s Skill) Option {
	return func(g *Game) { g.skill = s }
}
