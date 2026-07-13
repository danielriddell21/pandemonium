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
	ctx       *audio.Context
	players   map[iaudio.Cue]*audio.Player
	ambient   *audio.Player
	music     *audio.Player // looping musical bed, swapped as it darkens
	musicBand int           // the band the current music player was built for
	playing   bool          // whether the looping beds have been started
	depth     int           // levels reached; the ambient sinks as this grows
	listener  sim.Vec2      // the player's position, for distance attenuation

	sfxVolume     float64 // user volume scale for sound effects (0..1)
	ambientVolume float64 // user volume scale for the ambient loop (0..1)
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
	musicLevel   = 0.5  // music bed volume, before the user's ambient scale
)

// Compile-time check that Audio can drive the simulation's observations.
var _ sim.Observer = (*Audio)(nil)

// NewAudio builds the playback engine: one reusable player per sound effect plus a
// looping ambient track, all from the procedural PCM.
func NewAudio() (*Audio, error) {
	ctx := audio.NewContext(iaudio.SampleRate)
	a := &Audio{ctx: ctx, players: make(map[iaudio.Cue]*audio.Player), sfxVolume: 1, ambientVolume: 1}
	for cue, pcm := range iaudio.Synth() {
		p := ctx.NewPlayerFromBytes(pcm)
		p.SetVolume(0.6)
		a.players[cue] = p
	}
	ap, err := a.loopingPlayer(iaudio.Ambient())
	if err != nil {
		return nil, fmt.Errorf("audio: ambient player: %w", err)
	}
	a.ambient = ap
	mp, err := a.loopingPlayer(iaudio.Music(0))
	if err != nil {
		return nil, fmt.Errorf("audio: music player: %w", err)
	}
	a.music = mp
	a.applyAmbient()
	return a, nil
}

// loopingPlayer builds a seamlessly looping player from one PCM buffer.
func (a *Audio) loopingPlayer(pcm []byte) (*audio.Player, error) {
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
	return a.ctx.NewPlayer(loop)
}

// SetVolumes applies the user's sound-effect and ambient volume scales (each
// clamped to [0, 1]) on top of the engine's own levels.
func (a *Audio) SetVolumes(sfx, ambient float64) {
	if a == nil {
		return
	}
	a.sfxVolume = clamp01(sfx)
	a.ambientVolume = clamp01(ambient)
	a.applyAmbient()
}

func clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// StartAmbient begins the looping ambient track and the music bed.
func (a *Audio) StartAmbient() {
	if a == nil {
		return
	}
	a.playing = true
	if a.ambient != nil {
		a.ambient.Play()
	}
	if a.music != nil {
		a.music.Play()
	}
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

// deepen lowers the ambient volume one notch as the run reaches a new level and
// darkens the music bed when the run crosses into a new music band.
func (a *Audio) deepen() {
	a.depth++
	a.applyAmbient()
	a.updateMusicBand()
}

// updateMusicBand swaps the music bed for the current depth's darker variant
// when the band changes. The swap happens at a level boundary, where the tally
// screen masks any seam.
func (a *Audio) updateMusicBand() {
	band := iaudio.MusicBand(a.depth)
	if band == a.musicBand || a.music == nil {
		return
	}
	next, err := a.loopingPlayer(iaudio.Music(band))
	if err != nil {
		return // keep the current bed playing if the swap can't be built
	}
	old := a.music
	a.music = next
	a.musicBand = band
	a.applyAmbient()
	if a.playing {
		a.music.Play()
	}
	_ = old.Close()
}

// applyAmbient sets the ambient and music players' volumes from the depth curve
// scaled by the user's ambient volume.
func (a *Audio) applyAmbient() {
	if a.ambient != nil {
		vol := ambientLoud - ambientFade*float64(a.depth)
		if vol < ambientQuiet {
			vol = ambientQuiet
		}
		a.ambient.SetVolume(vol * a.ambientVolume)
	}
	if a.music != nil {
		a.music.SetVolume(musicLevel * a.ambientVolume)
	}
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
	p.SetVolume(0.6 * scale * a.sfxVolume)
	_ = p.Rewind()
	p.Play()
}
