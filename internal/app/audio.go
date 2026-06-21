package app

import (
	"bytes"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"

	iaudio "github.com/danielriddell21/pandemonium/internal/audio"
	"github.com/danielriddell21/pandemonium/internal/sim"
)

// Audio plays the procedurally synthesised sound effects and ambient loop through
// Ebiten. It is the only place sound touches Ebiten; the synthesis itself is the
// pure internal/audio package. Every method is nil-safe, so a game built without
// audio (e.g. on a machine with no sound device) simply runs silent.
type Audio struct {
	ctx      *audio.Context
	players  map[iaudio.Cue]*audio.Player
	ambient  *audio.Player
	depth    int      // levels reached; the ambient sinks as this grows
	listener sim.Vec2 // the player's position, for distance attenuation
}

// maxAudible is the distance (in tiles) beyond which a sound effect fades to
// nothing.
const maxAudible = 20.0

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

// SetListener records the player's position so subsequent sounds are attenuated
// by how far away they happen. The app sets it each tick before stepping the sim.
func (a *Audio) SetListener(pos sim.Vec2) {
	if a != nil {
		a.listener = pos
	}
}

// Observe plays the sound mapped to a simulation observation, if any, attenuated
// by its distance from the listener, and sinks the ambient a little each time a
// level is cleared. It lets the engine sit alongside telemetry as a second
// observer on the game.
func (a *Audio) Observe(o sim.Observation) {
	if a == nil {
		return
	}
	if o.Kind == sim.ObsExit {
		a.deepen()
	}
	if cue, ok := iaudio.CueFor(o.Kind); ok {
		event := sim.Vec2{X: float64(o.At.X) + 0.5, Y: float64(o.At.Y) + 0.5}
		a.playAt(cue, distanceVolume(a.listener, event))
	}
}

// distanceVolume falls from 1 at the listener to 0 at maxAudible.
func distanceVolume(listener, event sim.Vec2) float64 {
	dx, dy := event.X-listener.X, event.Y-listener.Y
	d := math.Hypot(dx, dy)
	if d >= maxAudible {
		return 0
	}
	return 1 - d/maxAudible
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
// observation, so it always plays at full volume.
func (a *Audio) Fire() { a.playAt(iaudio.CueFire, 1) }

// playAt restarts and triggers the player for a cue at the given volume scale.
func (a *Audio) playAt(cue iaudio.Cue, scale float64) {
	if a == nil || scale <= 0 {
		return
	}
	p := a.players[cue]
	if p == nil {
		return
	}
	p.SetVolume(0.6 * scale)
	_ = p.Rewind()
	p.Play()
}
