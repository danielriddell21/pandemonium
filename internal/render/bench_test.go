package render

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
	"github.com/danielriddell21/pandemonium/internal/world"
)

func benchLevel(b *testing.B, cfg world.Config) *sim.Game {
	b.Helper()
	l, err := world.Generate(cfg)
	if err != nil {
		b.Fatal(err)
	}
	return sim.New(l)
}

func BenchmarkFrame(b *testing.B) {
	g := benchLevel(b, world.Config{Width: 48, Height: 32, Seed: 7})
	r := NewRenderer(DefaultConfig())
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.Frame(g)
	}
}

func BenchmarkFrameArena(b *testing.B) {
	g := benchLevel(b, world.Config{Width: 48, Height: 32, Seed: 5, Arena: true})
	r := NewRenderer(DefaultConfig())
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.Frame(g)
	}
}
