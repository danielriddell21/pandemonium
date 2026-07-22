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
