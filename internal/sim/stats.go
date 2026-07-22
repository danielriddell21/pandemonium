package sim

type LevelStats struct {
	Kills, KillsTotal     int
	Items, ItemsTotal     int
	Secrets, SecretsTotal int
	Elapsed               float64
	Par                   float64
}

func (g *Game) LevelStats() LevelStats {
	return LevelStats{
		Kills:        g.kills,
		KillsTotal:   g.killsTotal,
		Items:        g.items,
		ItemsTotal:   g.itemsTotal,
		Secrets:      g.found,
		SecretsTotal: g.foundTotal,
		Elapsed:      g.elapsed,
		Par:          g.par,
	}
}

func (s LevelStats) KillsPct() int { return pct(s.Kills, s.KillsTotal) }

func (s LevelStats) ItemsPct() int { return pct(s.Items, s.ItemsTotal) }

func (s LevelStats) SecretsPct() int { return pct(s.Secrets, s.SecretsTotal) }

func pct(got, total int) int {
	if total <= 0 {
		return 100
	}
	return got * 100 / total
}

func parTime(steps int) float64 {
	const base, secPerTile = 12.0, 0.7
	if steps <= 0 {
		return base
	}
	return base + float64(steps)*secPerTile
}
