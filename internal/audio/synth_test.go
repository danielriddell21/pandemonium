package audio

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

func TestSynthProducesEveryCue(t *testing.T) {
	s := Synth()
	for _, cue := range []Cue{CueFire, CueHit, CueDoorOpen, CuePickup, CueSecret, CueDeath, CueExit, CueMenu} {
		pcm, ok := s[cue]
		if !ok || len(pcm) == 0 {
			t.Errorf("cue %d produced no PCM", cue)
			continue
		}
		if len(pcm)%bytesPerFrame != 0 {
			t.Errorf("cue %d PCM length %d is not a whole number of stereo frames", cue, len(pcm))
		}
	}
}

func TestAmbientIsWholeFrames(t *testing.T) {
	a := Ambient()
	if len(a) == 0 {
		t.Fatal("ambient produced no PCM")
	}
	if len(a)%bytesPerFrame != 0 {
		t.Errorf("ambient length %d is not a whole number of frames", len(a))
	}
}

func TestSynthIsDeterministic(t *testing.T) {
	a, b := Synth(), Synth()
	for cue, pa := range a {
		pb := b[cue]
		if len(pa) != len(pb) {
			t.Fatalf("cue %d length differs between builds", cue)
		}
		for i := range pa {
			if pa[i] != pb[i] {
				t.Fatalf("cue %d byte %d differs between builds", cue, i)
			}
		}
	}
}

func TestCueForMapping(t *testing.T) {
	cases := []struct {
		kind sim.ObservationKind
		cue  Cue
		ok   bool
	}{
		{sim.ObsKill, CueHit, true},
		{sim.ObsDoor, CueDoorOpen, true},
		{sim.ObsItem, CuePickup, true},
		{sim.ObsSecret, CueSecret, true},
		{sim.ObsDeath, CueDeath, true},
		{sim.ObsExit, CueExit, true},
		{sim.ObsMove, 0, false},
		{sim.ObsMarker, 0, false},
	}
	for _, c := range cases {
		got, ok := CueFor(c.kind)
		if ok != c.ok || (ok && got != c.cue) {
			t.Errorf("CueFor(%v) = (%v,%v), want (%v,%v)", c.kind, got, ok, c.cue, c.ok)
		}
	}
}
