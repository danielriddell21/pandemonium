// Package audio procedurally synthesises the game's sound effects and ambient
// pad as raw PCM. It is pure (no Ebiten, no files, deterministic) so it stays
// headless-testable; the Ebiten playback that turns this PCM into sound lives in
// the app layer.
package audio

import (
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/pandemonium/internal/sim"
)

// PCM format the synth emits: 16-bit little-endian, two channels (interleaved).
const (
	SampleRate      = 44100
	ChannelCount    = 2
	BitDepthInBytes = 2
	bytesPerFrame   = ChannelCount * BitDepthInBytes
)

// Cue identifies a synthesised sound.
type Cue int

// The set of synthesised sounds.
const (
	CueFire     Cue = iota // weapon discharge
	CueHit                 // a demon takes a lethal hit
	CueDoorOpen            // a door swings open
	CuePickup              // an item is collected
	CueSecret              // a secret area is found
	CueDeath               // the player dies
	CueExit                // the level exit is reached
	CueMenu                // a menu item is moved to or selected (UI blip)
)

// CueFor maps a simulation observation to the sound it should trigger, and
// reports false for observations that make no sound. Firing is player-driven and
// has no observation, so it is not mapped here.
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

// Synth builds every sound effect as PCM, keyed by cue. The result is
// deterministic: the same build always yields identical bytes.
func Synth() map[Cue][]byte {
	return map[Cue][]byte{
		CueFire:     synthFire(),
		CueHit:      synthHit(),
		CueDoorOpen: synthDoor(),
		CuePickup:   synthPickup(),
		CueSecret:   synthSecret(),
		CueDeath:    synthDeath(),
		CueExit:     synthExit(),
		CueMenu:     synthMenu(),
	}
}

// Ambient builds a seamlessly loopable low drone. Its length and frequencies are
// chosen so an integer number of cycles fits the buffer, so looping has no click.
func Ambient() []byte {
	return renderPCM(4.0, func(t float64) float64 {
		lfo := 0.6 + 0.4*math.Sin(2*math.Pi*0.5*t)
		v := 0.18*math.Sin(2*math.Pi*55*t) +
			0.12*math.Sin(2*math.Pi*82.5*t) +
			0.07*math.Sin(2*math.Pi*110*t)
		return v * lfo
	})
}

// maxMusicBand caps how dark the music gets, so detuning never drives a voice
// down to an inaudible or negative frequency.
const maxMusicBand = 8

// musicLoop is the music bed's loop length in seconds. Every voice frequency is
// an integer (cycles per second), so a whole number of cycles fits the loop and
// it repeats without a click.
const musicLoop = 8.0

// MusicBand maps how deep the run has gone to a music darkness band: the bed
// thins, slows and detunes a notch every few levels, then holds at the floor.
func MusicBand(depth int) int {
	b := depth / 3
	if b > maxMusicBand {
		b = maxMusicBand
	}
	return b
}

// Music builds a seamlessly loopable minor-mode pad for the given darkness band:
// a low triad pulsing under a sparse high shimmer. Deeper bands flatten the
// pitch, slow the pulse, drop the minor third and fade the shimmer, so the bed
// grows colder and emptier the further in the run goes. Deterministic.
func Music(band int) []byte {
	if band < 0 {
		band = 0
	} else if band > maxMusicBand {
		band = maxMusicBand
	}
	detune := float64(band)                // flatten every voice by ~1 Hz per band
	voices := []float64{55, 82, 110}       // A1 root, E2 fifth, A2 octave
	const third = 65                       // C2 minor third, present early, gone deep
	pulses := math.Max(2, 8-float64(band)) // slows with depth, stays integer
	shimmer := 220 - 2*detune              // a high voice, also flattening
	shimmerAmp := math.Max(0, 1-0.3*float64(band))
	thirdAmp := math.Max(0, 1-0.2*float64(band))

	return renderPCM(musicLoop, func(t float64) float64 {
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

// renderPCM samples gen over dur seconds and encodes it to interleaved stereo
// 16-bit little-endian PCM. gen returns a sample in [-1, 1] for time t.
func renderPCM(dur float64, gen func(t float64) float64) []byte {
	n := int(dur * SampleRate)
	buf := make([]byte, n*bytesPerFrame)
	for i := range n {
		v := gen(float64(i) / SampleRate)
		if v > 1 {
			v = 1
		} else if v < -1 {
			v = -1
		}
		s := int16(v * 32767)
		lo, hi := byte(s), byte(s>>8)
		off := i * bytesPerFrame
		buf[off], buf[off+1] = lo, hi   // left
		buf[off+2], buf[off+3] = lo, hi // right
	}
	return buf
}

func env(t, decay float64) float64 { return math.Exp(-t * decay) }

func synthFire() []byte {
	r := rand.New(rand.NewPCG(1, 2))
	return renderPCM(0.12, func(t float64) float64 {
		e := env(t, 40)
		noise := r.Float64()*2 - 1
		chirp := math.Sin(2 * math.Pi * (520 - 1500*t) * t)
		return (0.7*noise + 0.5*chirp) * e
	})
}

func synthHit() []byte {
	r := rand.New(rand.NewPCG(3, 4))
	return renderPCM(0.10, func(t float64) float64 {
		e := env(t, 55)
		thud := math.Sin(2 * math.Pi * 150 * t)
		noise := r.Float64()*2 - 1
		return (0.6*thud + 0.4*noise) * e
	})
}

func synthDoor() []byte {
	r := rand.New(rand.NewPCG(5, 6))
	return renderPCM(0.40, func(t float64) float64 {
		e := math.Min(1, t*8) * env(t, 5) // brief rise, slow fall
		rumble := math.Sin(2 * math.Pi * (90 + 40*t) * t)
		grit := (r.Float64()*2 - 1) * 0.2
		return (0.55*rumble + grit) * e
	})
}

func synthPickup() []byte {
	return renderPCM(0.18, func(t float64) float64 {
		f := 660.0
		local := t
		if t > 0.08 {
			f, local = 990.0, t-0.08
		}
		return 0.5 * math.Sin(2*math.Pi*f*t) * env(local, 22)
	})
}

func synthSecret() []byte {
	notes := []float64{523, 659, 784} // C5, E5, G5 arpeggio
	return renderPCM(0.45, func(t float64) float64 {
		idx := int(t / 0.15)
		if idx >= len(notes) {
			idx = len(notes) - 1
		}
		local := t - float64(idx)*0.15
		return 0.45 * math.Sin(2*math.Pi*notes[idx]*t) * env(local, 10)
	})
}

func synthDeath() []byte {
	return renderPCM(0.6, func(t float64) float64 {
		f := 300 - 360*t // slides down toward ~80Hz
		if f < 60 {
			f = 60
		}
		return 0.5 * math.Sin(2*math.Pi*f*t) * env(t, 4)
	})
}

// synthMenu is a short, clean UI blip for moving through and selecting menu items.
func synthMenu() []byte {
	return renderPCM(0.05, func(t float64) float64 {
		return 0.4 * math.Sin(2*math.Pi*880*t) * env(t, 60)
	})
}

func synthExit() []byte {
	return renderPCM(0.5, func(t float64) float64 {
		base := 0.4 * math.Sin(2*math.Pi*(330+220*t)*t)
		fifth := 0.25 * math.Sin(2*math.Pi*(495+330*t)*t)
		return (base + fifth) * math.Min(1, t*6) * env(t, 3)
	})
}
