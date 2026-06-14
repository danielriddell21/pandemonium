package sim

// LevelStats is an end-of-level summary of how thoroughly the player cleared a
// level: demons slain, items gathered, secrets uncovered, and time taken. It
// feeds the intermission screen shown between levels.
type LevelStats struct {
	Kills, KillsTotal     int
	Items, ItemsTotal     int
	Secrets, SecretsTotal int
	Elapsed               float64 // seconds spent on the level
}

// LevelStats snapshots the current level's tallies.
func (g *Game) LevelStats() LevelStats {
	return LevelStats{
		Kills:        g.kills,
		KillsTotal:   g.killsTotal,
		Items:        g.items,
		ItemsTotal:   g.itemsTotal,
		Secrets:      g.found,
		SecretsTotal: g.foundTotal,
		Elapsed:      g.elapsed,
	}
}

// KillsPct gives kills as a whole-number percentage, treating "none to find" as
// a full 100%.
func (s LevelStats) KillsPct() int { return pct(s.Kills, s.KillsTotal) }

// ItemsPct gives items collected as a whole-number percentage (100% if none).
func (s LevelStats) ItemsPct() int { return pct(s.Items, s.ItemsTotal) }

// SecretsPct gives secrets found as a whole-number percentage (100% if none).
func (s LevelStats) SecretsPct() int { return pct(s.Secrets, s.SecretsTotal) }

func pct(got, total int) int {
	if total <= 0 {
		return 100
	}
	return got * 100 / total
}
