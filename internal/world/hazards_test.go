package world

import "testing"

func TestHazardsAreWalkableAndClearOfSpawn(t *testing.T) {
	found := false
	for seed := int64(0); seed < 60 && !found; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if len(l.Hazard) == 0 {
			continue
		}
		found = true
		for c, hz := range l.Hazard {
			if hz.Rate <= 0 {
				t.Errorf("seed %d: hazard %v has non-positive rate %v", seed, c, hz.Rate)
			}
			if !l.At(c.X, c.Y).Walkable() {
				t.Errorf("seed %d: hazard %v is not walkable", seed, c)
			}
			if cheby(c, l.Spawn) < 4 {
				t.Errorf("seed %d: hazard %v too close to spawn", seed, c)
			}
		}
	}
	if !found {
		t.Skip("no hazard pool generated in scanned seeds")
	}
}

func TestHazardsDeterministic(t *testing.T) {
	cfg := Config{Width: 48, Height: 32, Seed: 23}
	a, _ := Generate(cfg)
	b, _ := Generate(cfg)
	if len(a.Hazard) != len(b.Hazard) {
		t.Fatalf("hazard count differs: %d vs %d", len(a.Hazard), len(b.Hazard))
	}
	for c, r := range a.Hazard {
		if b.Hazard[c] != r {
			t.Errorf("hazard at %v differs: %v vs %v", c, r, b.Hazard[c])
		}
	}
}
