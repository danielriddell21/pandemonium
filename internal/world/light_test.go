package world

import "testing"

func TestLightVariesAndSpawnIsBright(t *testing.T) {
	dimSeen := false
	for seed := int64(0); seed < 40; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		if l.LightAt(l.Spawn.X, l.Spawn.Y) != 1.0 {
			t.Errorf("seed %d: spawn should be fully lit, got %v", seed, l.LightAt(l.Spawn.X, l.Spawn.Y))
		}
		for _, v := range l.Light {
			if v < 1.0 {
				dimSeen = true
			}
			if v <= 0 || v > 1.0 {
				t.Fatalf("seed %d: light %v out of (0,1]", seed, v)
			}
		}
	}
	if !dimSeen {
		t.Error("expected some rooms to be dimmer than full light")
	}
}

func TestLightDeterministic(t *testing.T) {
	cfg := Config{Width: 48, Height: 32, Seed: 31}
	a, _ := Generate(cfg)
	b, _ := Generate(cfg)
	for i := range a.Light {
		if a.Light[i] != b.Light[i] {
			t.Fatalf("light differs at %d between identical seeds", i)
		}
	}
}
