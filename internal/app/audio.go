package app

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/audio"

	iaudio "github.com/danielriddell21/pandemonium/internal/audio"
	"github.com/danielriddell21/pandemonium/internal/sim"
)

// Audio plays the procedurally synthesised sound effects and ambient loop through
// Ebiten. It is the only place sound touches Ebiten; the synthesis itself is the
// pure internal/audio package. Every method is nil-safe, so a game built without
// audio (e.g. on a machine with no sound device) simply runs silent.
type Audio struct {
	ctx     *audio.Context
	players map[iaudio.Cue]*audio.Player
	ambient *audio.Player
	depth   int // levels reached; the ambient sinks as this grows
}

// ambient volume settles from ambientLoud toward ambientQuiet as the run deepens,
// so the soundscape grows colder the further in you get.
const (
	ambientLoud  = 0.35
	ambientQuiet = 0.12
	ambientFade  = 0.03 // volume lost per level cleared
)

// Compile-time check that Audio can drive the simulation's observations.
var _ sim.Observer = (*Audio)(nil)

// NewAudio builds the playback engine: one reusable player per sound effect plus a
// looping ambient track, all from the procedural PCM.
func NewAudio() (*Audio, error) {
	ctx := audio.NewContext(iaudio.SampleRate)
	a := &Audio{ctx: ctx, players: make(map[iaudio.Cue]*audio.Player)}
	for cue, pcm := range iaudio.Synth() {
		p := ctx.NewPlayerFromBytes(pcm)
		p.SetVolume(0.6)
		a.players[cue] = p
	}
	amb := iaudio.Ambient()
	loop := audio.NewInfiniteLoop(bytes.NewReader(amb), int64(len(amb)))
	ap, err := ctx.NewPlayer(loop)
	if err != nil {
		return nil, fmt.Errorf("audio: ambient player: %w", err)
	}
	ap.SetVolume(ambientLoud)
	a.ambient = ap
	return a, nil
}

// StartAmbient begins the looping ambient track.
func (a *Audio) StartAmbient() {
	if a == nil || a.ambient == nil {
		return
	}
	a.ambient.Play()
}

// Observe plays the sound mapped to a simulation observation, if any, and sinks
// the ambient a little each time a level is cleared. It lets the engine sit
// alongside telemetry as a second observer on the game.
func (a *Audio) Observe(o sim.Observation) {
	if a == nil {
		return
	}
	if o.Kind == sim.ObsExit {
		a.deepen()
	}
	if cue, ok := iaudio.CueFor(o.Kind); ok {
		a.play(cue)
	}
}

// deepen lowers the ambient volume one notch as the run reaches a new level.
func (a *Audio) deepen() {
	a.depth++
	if a.ambient == nil {
		return
	}
	vol := ambientLoud - ambientFade*float64(a.depth)
	if vol < ambientQuiet {
		vol = ambientQuiet
	}
	a.ambient.SetVolume(vol)
}

// Fire plays the weapon-discharge sound, which is player-driven rather than an
// observation.
func (a *Audio) Fire() { a.play(iaudio.CueFire) }

// play restarts and triggers the player for a cue.
func (a *Audio) play(cue iaudio.Cue) {
	if a == nil {
		return
	}
	p := a.players[cue]
	if p == nil {
		return
	}
	_ = p.Rewind()
	p.Play()
}
