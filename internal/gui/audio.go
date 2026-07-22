package gui

import (
	"bytes"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"

	iaudio "github.com/danielriddell21/pandemonium/internal/audio"
	"github.com/danielriddell21/pandemonium/internal/sim"
)

type Audio struct {
	ctx       *audio.Context
	players   map[iaudio.Cue]*audio.Player
	ambient   *audio.Player
	music     *audio.Player
	musicBand int
	playing   bool
	depth     int
	listener  sim.Vec2

	sfxVolume     float64
	ambientVolume float64
}

const maxAudible = 20.0

const (
	ambientLoud  = 0.35
	ambientQuiet = 0.12
	ambientFade  = 0.03
	musicLevel   = 0.5
)

var _ sim.Observer = (*Audio)(nil)

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

func (a *Audio) loopingPlayer(pcm []byte) (*audio.Player, error) {
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
	p, err := a.ctx.NewPlayer(loop)
	if err != nil {
		return nil, fmt.Errorf("%w: create looping player: %w", ErrAudioUnavailable, err)
	}
	return p, nil
}

func (a *Audio) SetVolumes(sfx, ambient float64) {
	if a == nil {
		return
	}
	a.sfxVolume = clamp01(sfx)
	a.ambientVolume = clamp01(ambient)
	a.applyAmbient()
}

func clamp01(v float64) float64 { return max(0, min(1, v)) }

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

func (a *Audio) SetListener(pos sim.Vec2) {
	if a != nil {
		a.listener = pos
	}
}

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

func distanceVolume(listener, event sim.Vec2) float64 {
	dx, dy := event.X-listener.X, event.Y-listener.Y
	d := math.Hypot(dx, dy)
	if d >= maxAudible {
		return 0
	}
	return 1 - d/maxAudible
}

func (a *Audio) deepen() {
	a.depth++
	a.applyAmbient()
	a.updateMusicBand()
}

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

func (a *Audio) Fire() { a.playAt(iaudio.CueFire, 1) }

func (a *Audio) Menu() { a.playAt(iaudio.CueMenu, 1) }

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
