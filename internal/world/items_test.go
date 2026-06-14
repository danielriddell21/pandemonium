package world

import "testing"

func TestGenerateScattersItems(t *testing.T) {
	l, err := Generate(Config{Width: 48, Height: 32, Seed: 7})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(l.Items) == 0 {
		t.Fatal("expected the level to hold some items")
	}

	seen := make(map[Coord]bool)
	for _, it := range l.Items {
		if !l.At(it.At.X, it.At.Y).Walkable() {
			t.Errorf("item %v sits on a non-walkable tile", it)
		}
		if it.At == l.Spawn {
			t.Errorf("item placed on the spawn at %v", it.At)
		}
		if cheby(it.At, l.Spawn) < 3 {
			t.Errorf("item %v placed inside the spawn area", it.At)
		}
		if seen[it.At] {
			t.Errorf("two items share cell %v", it.At)
		}
		seen[it.At] = true
	}
}

func TestItemPlacementIsDeterministic(t *testing.T) {
	cfg := Config{Width: 48, Height: 32, Seed: 99}
	a, err := Generate(cfg)
	if err != nil {
		t.Fatalf("generate a: %v", err)
	}
	b, err := Generate(cfg)
	if err != nil {
		t.Fatalf("generate b: %v", err)
	}
	if len(a.Items) != len(b.Items) {
		t.Fatalf("item count differs: %d vs %d", len(a.Items), len(b.Items))
	}
	for i := range a.Items {
		if a.Items[i] != b.Items[i] {
			t.Errorf("item %d differs: %v vs %v", i, a.Items[i], b.Items[i])
		}
	}
}
