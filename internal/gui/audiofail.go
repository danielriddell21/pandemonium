package gui

import (
	"errors"
	"fmt"
	"os"
)

var ErrAudioUnavailable = errors.New("audio device unavailable")

func (g *Game) handleRunError(err error) error {
	if !errors.Is(err, ErrAudioUnavailable) {
		return err
	}
	fmt.Fprintln(os.Stderr, "audio device unavailable — disabling sound; relaunch to play")
	g.settings.Sound = false
	_ = g.settings.Save() // best effort; the in-memory value still takes effect next time
	return nil
}
