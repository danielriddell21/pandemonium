package world

import "testing"

func TestSkyAppearsAndIsWellFormed(t *testing.T) {
	skySeen := false
	for seed := int64(0); seed < 40; seed++ {
		l, err := Generate(Config{Width: 48, Height: 32, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		for y := range l.Height {
			for x := range l.Width {
				if !l.SkyAt(x, y) {
					continue
				}
				skySeen = true
				if !l.At(x, y).Walkable() {
					t.Fatalf("seed %d: sky on a non-walkable cell %d,%d", seed, x, y)
				}
				if l.LightAt(x, y) != 1 {
					t.Errorf("seed %d: sky cell %d,%d not fully lit: %v", seed, x, y, l.LightAt(x, y))
				}
				if h := l.Ceil(x, y) - l.Floor(x, y); h < skyHeadroom-1e-9 {
					t.Errorf("seed %d: sky cell %d,%d too low: headroom %.2f", seed, x, y, h)
				}
			}
		}
	}
	if !skySeen {
		t.Error("expected some open-air (sky) cells across the seeds")
	}
}

func TestSkyDeterministic(t *testing.T) {
	cfg := Config{Width: 48, Height: 32, Seed: 19}
	a, _ := Generate(cfg)
	b, _ := Generate(cfg)
	for i := range a.Sky {
		if a.Sky[i] != b.Sky[i] {
			t.Fatalf("sky differs at %d between identical seeds", i)
		}
	}
}

func TestSkyAtNilSafe(t *testing.T) {
	l := &Level{Width: 4, Height: 4} // no Sky layer allocated
	if l.SkyAt(1, 1) {
		t.Error("a level with no sky layer should report no sky")
	}
}
