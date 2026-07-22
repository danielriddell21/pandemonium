package audio

import (
	"math"

	"github.com/danielriddell21/crucible/synth"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

const (
	// SampleRate is the audio player's sample rate, re-exported from
	// crucible/synth for the gui player.
	SampleRate = synth.SampleRate
	// bytesPerFrame is the size of one stereo 16-bit frame.
	bytesPerFrame = synth.BytesPerFrame
)

type Cue int

const (
	CueFire Cue = iota
	CueHit
	CueDoorOpen
	CuePickup
	CueSecret
	CueDeath
	CueExit
	CueMenu
)

func CueFor(k sim.ObservationKind) (Cue, bool) {
	switch k {
	case sim.ObsKill:
		return CueHit, true
	case sim.ObsDoor:
		return CueDoorOpen, true
	case sim.ObsItem:
		return CuePickup, true
	case sim.ObsSecret:
		return CueSecret, true
	case sim.ObsDeath:
		return CueDeath, true
	case sim.ObsExit:
		return CueExit, true
	default:
		return 0, false
	}
}

func Synth() map[Cue][]byte {
	return map[Cue][]byte{
		CueFire:     synthFire(),
		CueHit:      synth.Thud(150, 0.10, 55, 3, 4),
		CueDoorOpen: synth.Rumble(90, 40, 0.40, 5, 6),
		CuePickup:   synth.TwoTone(660, 990, 0.08, 0.18, 22),
		CueSecret:   synth.Arpeggio([]float64{523, 659, 784}, 0.15, 0.45, 10),
		CueDeath:    synth.Slide(300, -360, 60, 0.6, 4),
		CueExit:     synthExit(),
		CueMenu:     synth.Blip(880, 0.05, 60),
	}
}

// synthFire is a noise-and-chirp burst with no shared shape; it stays local
// on the crucible primitives.
func synthFire() []byte {
	noise := synth.Noise(1, 2)
	return synth.Render(0.12, func(t float64) float64 {
		e := synth.Env(t, 40)
		chirp := synth.Sine(520-1500*t, t)
		return (0.7*noise() + 0.5*chirp) * e
	})
}

// synthExit is a rising perfect-fifth swell; it stays local on the crucible
// primitives.
func synthExit() []byte {
	return synth.Render(0.5, func(t float64) float64 {
		base := 0.4 * synth.Sine(330+220*t, t)
		fifth := 0.25 * synth.Sine(495+330*t, t)
		return (base + fifth) * synth.Attack(t, 6) * synth.Env(t, 3)
	})
}

func Ambient() []byte {
	return synth.Render(4.0, func(t float64) float64 {
		lfo := 0.6 + 0.4*math.Sin(2*math.Pi*0.5*t)
		v := 0.18*math.Sin(2*math.Pi*55*t) +
			0.12*math.Sin(2*math.Pi*82.5*t) +
			0.07*math.Sin(2*math.Pi*110*t)
		return v * lfo
	})
}

const maxMusicBand = 8

const musicLoop = 8.0

func MusicBand(depth int) int {
	b := depth / 3
	if b > maxMusicBand {
		b = maxMusicBand
	}
	return b
}

func Music(band int) []byte {
	if band < 0 {
		band = 0
	} else if band > maxMusicBand {
		band = maxMusicBand
	}
	detune := float64(band)           // flatten every voice by ~1 Hz per band
	voices := []float64{55, 82, 110}  // A1 root, E2 fifth, A2 octave
	const third = 65                  // C2 minor third, present early, gone deep
	pulses := max(2, 8-float64(band)) // slows with depth, stays integer
	shimmer := 220 - 2*detune         // a high voice, also flattening
	shimmerAmp := max(0, 1-0.3*float64(band))
	thirdAmp := max(0, 1-0.2*float64(band))

	return synth.Render(musicLoop, func(t float64) float64 {
		// Amplitude pulse: an integer number of cycles over the loop.
		pulse := 0.55 + 0.45*math.Sin(2*math.Pi*(pulses/musicLoop)*t-math.Pi/2)
		low := 0.0
		for _, f := range voices {
			low += math.Sin(2 * math.Pi * (f - detune) * t)
		}
		low *= 0.10
		low += 0.06 * thirdAmp * math.Sin(2*math.Pi*(third-detune)*t)
		// A single slow swell over the loop carries the shimmer in and out.
		swell := 0.5 + 0.5*math.Sin(2*math.Pi*(1/musicLoop)*t-math.Pi/2)
		hi := shimmerAmp * 0.05 * swell * math.Sin(2*math.Pi*shimmer*t)
		return low*pulse + hi
	})
}
