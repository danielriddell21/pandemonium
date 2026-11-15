package gui

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrAudioUnavailable = errors.New("audio device unavailable")

func (g *Game) handleRunError(err error) error {
	if !audioFailure(err) {
		return err
	}
	fmt.Fprintln(os.Stderr, "audio device unavailable — disabling sound; relaunch to play")
	g.settings.Sound = false
	_ = g.settings.Save() // best effort; the in-memory value still takes effect next time
	return nil
}

const engineAudioError = "audio: audio error"

func audioFailure(err error) bool {
	if err == nil {
		return false
	}
	// A machine with no sound device fails inside the engine's audio thread,
	// after the players were built: that error carries no sentinel, only the
	// engine's own prefix.
	return errors.Is(err, ErrAudioUnavailable) || strings.Contains(err.Error(), engineAudioError)
}
