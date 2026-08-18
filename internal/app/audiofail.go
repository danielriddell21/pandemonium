package app

import (
	"fmt"
	"os"
	"strings"
)

// isAudioError reports whether err comes from the audio device failing to open —
// the oto/ALSA failure Ebiten surfaces from the game loop on a machine with no
// sound device. Ebiten exposes no synchronous device check, so this classifies
// the error after the fact by its message.
func isAudioError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "oto") || strings.Contains(msg, "audio")
}

// handleRunError turns a fatal audio-device error into graceful degradation:
// rather than crashing, it disables sound for next time and exits cleanly, so a
// machine with no audio device self-heals after one relaunch. Any other error is
// returned unchanged.
func (g *Game) handleRunError(err error) error {
	if !isAudioError(err) {
		return err
	}
	fmt.Fprintln(os.Stderr, "audio device unavailable — disabling sound; relaunch to play")
	g.settings.Sound = false
	_ = g.settings.Save() // best effort; the in-memory value still takes effect next time
	return nil
}
