package app

import (
	"errors"
	"io"
	"testing"
)

func TestIsAudioError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"real oto/ALSA", errors.New("audio: audio error: oto: ALSA error at snd_pcm_open: \"default\": No such file or directory"), true},
		{"bare oto", errors.New("oto: device closed"), true},
		{"eof", io.EOF, false},
		{"unrelated", errors.New("renderer: out of memory"), false},
	}
	for _, c := range cases {
		if got := isAudioError(c.err); got != c.want {
			t.Errorf("%s: isAudioError = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestHandleRunErrorSwallowsAudio(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // keep the best-effort Save off the real config dir
	g := &Game{settings: DefaultSettings()}
	audioErr := errors.New("audio: audio error: oto: ALSA error at snd_pcm_open")
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
