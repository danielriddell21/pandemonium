package world

import "testing"

func TestSecretsCarryRewards(t *testing.T) {
	found := false
	for seed := int64(0); seed < 100; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if len(l.Secrets) == 0 {
			continue
		}
		found = true
		for _, s := range l.Secrets {
			if !l.At(s.X, s.Y).Walkable() {
				t.Errorf("seed %d: secret %v is not walkable", seed, s)
			}
			if !itemAt(l, s) {
				t.Errorf("seed %d: secret %v has no reward item", seed, s)
			}
		}
	}
	if !found {
		t.Skip("no secrets generated in scanned seeds")
	}
}

func itemAt(l *Level, c Coord) bool {
	for _, it := range l.Items {
		if it.At == c {
			return true
		}
	}
	return false
}
