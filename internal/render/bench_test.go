package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

// Perf budget: the app advances sim + render once per tick at 60 TPS, so the
// whole tick has ~16.6 ms. Frame should sit comfortably within that at the
// default 640x400 — measured around 6 ms/frame (ordinary and arena levels alike)
// with only ~21 allocations, leaving ample headroom. These benchmarks are a
// measuring tool, not a CI gate — timing is environment-dependent. Run them with
// `just bench` (or `go test -run=^$ -bench=. ./internal/render`).

func benchLevel(b *testing.B, cfg world.Config) *sim.Game {
	b.Helper()
	l, err := world.Generate(cfg)
	if err != nil {
		b.Fatal(err)
	}
	return sim.New(l)
}

// BenchmarkFrame measures a full frame of an ordinary generated level.
func BenchmarkFrame(b *testing.B) {
	g := benchLevel(b, world.Config{Width: 48, Height: 32, Seed: 7})
	r := NewRenderer(DefaultConfig())
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.Frame(g)
	}
}

// BenchmarkFrameArena measures the heavier case: an open, sky-lit arena where the
// boundary walk reaches far and many sprites are in view.
func BenchmarkFrameArena(b *testing.B) {
	g := benchLevel(b, world.Config{Width: 48, Height: 32, Seed: 5, Arena: true})
	r := NewRenderer(DefaultConfig())
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.Frame(g)
	}
}
