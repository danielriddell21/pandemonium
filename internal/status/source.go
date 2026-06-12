// Package status turns playtest telemetry into short HUD messages. It subscribes
// to the telemetry bus and posts lines to a shared overlay, choosing what to say
// through a swappable Source so the message policy can evolve independently of the
// wiring.
package status

import "github.com/danielriddell21/pandemonium/internal/hud"

// CueKind identifies what the player just did that might warrant a message.
type CueKind uint8

const (
	// CueExit is reaching the level exit.
	CueExit CueKind = iota
	// CueWrongDoor is opening a door that leads only to a dead end.
	CueWrongDoor
	// CueDecoy is reaching a tile that resembles the exit but is not it.
	CueDecoy
	// CueFork is crossing a junction.
	CueFork
)

// Cue is the context a Source uses to choose a line.
type Cue struct {
	Kind          CueKind
	Level         int  // zero-based level index
	OptimalChoice bool // for forks: the branch nearest the exit was taken
}

// Line is a chosen message with its presentation channel and lifetime in frames.
type Line struct {
	Text    string
	Channel hud.Channel
	Frames  int
}

// Source decides what to say for a cue and delivers it through emit. emit may be
// called synchronously, or later from another goroutine; it may be called zero or
// one time. Returning without calling emit stays silent. This async shape lets a
// future source that fetches lines out of process plug in without changing callers.
type Source interface {
	Request(cue Cue, emit func(Line))
}

// messageFrames is how long a posted line stays up (~2.5s at 60 fps).
const messageFrames = 150

// tableSource is the default scripted policy. It stays silent for the first
// couple of levels, surfaces diagnostic readouts (debug-only) in the mid band,
// and player-facing notices later.
type tableSource struct{}

// NewTableSource returns the default scripted message source.
func NewTableSource() Source { return tableSource{} }

func band(level int) int {
	switch {
	case level < 2:
		return 0
	case level < 5:
		return 1
	default:
		return 2
	}
}

func (tableSource) Request(c Cue, emit func(Line)) {
	if line, ok := scriptedLine(c); ok {
		emit(line)
	}
}

// scriptedLine computes the banded line for a cue, or reports false to stay silent.
func scriptedLine(c Cue) (Line, bool) {
	switch band(c.Level) {
	case 0:
		return Line{}, false
	case 1:
		if text, ok := diagnosticText(c); ok {
			return Line{Text: text, Channel: hud.Diagnostic, Frames: messageFrames}, true
		}
	default:
		if text, ok := noticeText(c); ok {
			return Line{Text: text, Channel: hud.Notice, Frames: messageFrames}, true
		}
	}
	return Line{}, false
}

// diagnosticText is the dev/playtest readout for a cue.
func diagnosticText(c Cue) (string, bool) {
	switch c.Kind {
	case CueExit:
		return "telemetry: level complete", true
	case CueWrongDoor:
		return "telemetry: dead-end door opened", true
	case CueDecoy:
		return "telemetry: decoy marker reached", true
	case CueFork:
		if c.OptimalChoice {
			return "telemetry: optimal branch taken", true
		}
		return "telemetry: suboptimal branch taken", true
	}
	return "", false
}

// noticeText is the player-facing line for a cue. Forks are intentionally silent
// to avoid chatter.
func noticeText(c Cue) (string, bool) {
	switch c.Kind {
	case CueExit:
		return "The exit. Naturally.", true
	case CueWrongDoor:
		return "Nothing behind that one.", true
	case CueDecoy:
		return "Not every door leads onward.", true
	}
	return "", false
}
