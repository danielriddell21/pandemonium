package gui

import (
	"errors"
	"fmt"
	"testing"
)

func TestHandleRunErrorSwallowsAudio(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // keep the best-effort Save off the real config dir
	g := &Game{settings: DefaultSettings()}
	audioErr := fmt.Errorf("audio: music player: %w: create looping player: %w",
		ErrAudioUnavailable, errors.New("oto: device closed"))
	if err := g.handleRunError(audioErr); err != nil {
		t.Errorf("audio error should be swallowed, got %v", err)
	}
	if g.settings.Sound {
		t.Error("an audio failure should disable sound for next time")
	}
}

func TestHandleRunErrorPassesThrough(t *testing.T) {
	g := &Game{settings: DefaultSettings()}
	boom := errors.New("something genuinely broke")
	if err := g.handleRunError(boom); !errors.Is(err, boom) {
		t.Errorf("non-audio error should pass through, got %v", err)
	}
	if !g.settings.Sound {
		t.Error("a non-audio error must not touch the sound setting")
	}
}

func TestHandleRunErrorNilIsNil(t *testing.T) {
	g := &Game{settings: DefaultSettings()}
	if err := g.handleRunError(nil); err != nil {
		t.Errorf("nil should stay nil, got %v", err)
	}
}

func TestHandleRunErrorSwallowsTheEnginesOwnAudioFailure(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	g := &Game{settings: DefaultSettings()}
	// What ebiten reports on a machine with no sound device at all.
	engine := fmt.Errorf("run game: %w", errors.New("audio: audio error: oto: initialization failed: "+
		"ALSA: oto: ALSA device initialization failed"))
	if err := g.handleRunError(engine); err != nil {
		t.Errorf("a missing sound device should not be fatal, got %v", err)
	}
	if g.settings.Sound {
		t.Error("a missing sound device should disable sound for next time")
	}
}

func TestAudioFailure(t *testing.T) {
	if audioFailure(nil) || audioFailure(errors.New("window: no display")) {
		t.Error("only audio errors count as audio failures")
	}
	if !audioFailure(ErrAudioUnavailable) {
		t.Error("the sentinel is an audio failure")
	}
}
