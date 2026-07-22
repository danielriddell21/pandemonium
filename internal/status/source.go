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
	// CueKill is felling a demon (throttled to the first of each level).
	CueKill
	// CueItem is collecting something (throttled to the first of each level).
	CueItem
	// CueSecret is uncovering a hidden area.
	CueSecret
	// CueDeath is dying and respawning.
	CueDeath
)

// Cue is the context a Source uses to choose a line: the immediate trigger plus a
// snapshot of how the run has gone so far.
type Cue struct {
	Kind          CueKind
	Level         int  // zero-based level index
	OptimalChoice bool // for forks: the branch nearest the exit was taken

	// Run-so-far context, drawn from the cumulative profile.
	LevelsCleared   int
	Deaths          int
	WrongDoorsTotal int
	ExploreScore    float64 // 0..1; lower means more retreading
	Rushing         bool    // tends to beeline rather than explore (low explore score)

	// Arrival marks the single moment the run first reaches the deepest band.
	Arrival bool
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

// deepestBand is the highest progression band, where the run's own length becomes
// the subject and the voice finally falls quiet.
const deepestBand = 4

// tableSource is the default scripted policy. It stays silent for the first
// couple of levels, surfaces diagnostic readouts (debug-only) in the mid band,
// and player-facing notices later.
type tableSource struct{}

// NewTableSource returns the default scripted message source.
func NewTableSource() Source { return tableSource{} }

func band(level int) int {
	switch {
	case level < 2:
		return 0 // silent
	case level < 5:
		return 1 // diagnostic (debug-only)
	case level < 9:
		return 2 // notice
	case level < 15:
		return 3 // notice, more pointed
	default:
		return 4 // the deepest register: the run's own length is the subject
	}
}

func (tableSource) Request(c Cue, emit func(Line)) {
	if line, ok := scriptedLine(c); ok {
		emit(line)
	}
}

// scriptedLine computes the banded line for a cue, or reports false to stay silent.
func scriptedLine(c Cue) (Line, bool) {
	if c.Arrival {
		// The single line for first reaching the deepest band.
		return Line{Text: "You've gone as deep as it goes. It just keeps going.", Channel: hud.Notice, Frames: messageFrames}, true
	}
	b := band(c.Level)
	switch b {
	case 0:
		// Diagnostic telemetry readouts surface from the very first level (they
		// only display when debug messages are enabled); no player-facing notices
		// yet — the drift stays silent until its later bands.
		if text, ok := diagnosticText(c); ok {
			return Line{Text: text, Channel: hud.Diagnostic, Frames: messageFrames}, true
		}
	case 1:
		// The mid band is mostly debug-only readouts, but a couple of impactful
		// moments leak a quiet player-facing whisper, so the drift is felt early.
		if text, ok := band1Whisper(c); ok {
			return Line{Text: text, Channel: hud.Notice, Frames: messageFrames}, true
		}
		if text, ok := diagnosticText(c); ok {
			return Line{Text: text, Channel: hud.Diagnostic, Frames: messageFrames}, true
		}
	default:
		if text, ok := noticeText(c, b); ok {
			return Line{Text: text, Channel: hud.Notice, Frames: messageFrames}, true
		}
	}
	return Line{}, false
}

// band1Whisper is the rare early player-facing line for the most charged moments,
// surfacing before the notice band proper. Everything else stays diagnostic here.
func band1Whisper(c Cue) (string, bool) {
	switch c.Kind {
	case CueDeath:
		return "Hm. Again.", true
	case CueSecret:
		return "Something tucked away.", true
	}
	return "", false
}

// diagnosticText is the dev/playtest readout for a cue, including a little run
// context where it is informative.
func diagnosticText(c Cue) (string, bool) {
	switch c.Kind {
	case CueExit:
		if c.Rushing {
			return "telemetry: level complete (rush pattern)", true
		}
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
	case CueKill:
		return "telemetry: hostile neutralised", true
	case CueItem:
		return "telemetry: item acquired", true
	case CueSecret:
		return "telemetry: hidden area logged", true
	case CueDeath:
		return "telemetry: respawn event recorded", true
	}
	return "", false
}

// noticeText is the player-facing line for a cue. It reacts to the run so far and
// reads more pointed in the later band. Forks are intentionally silent to avoid
// chatter.
func noticeText(c Cue, b int) (string, bool) {
	switch c.Kind {
	case CueExit:
		switch {
		case b >= 4:
			return "Another way out. Into another room exactly like this one.", true
		case c.Rushing:
			return "Straight to the exit. Predictable.", true
		case b >= 3:
			return "Another exit. You do keep finding them.", true
		default:
			return "The exit. Naturally.", true
		}
	case CueWrongDoor:
		switch {
		case b >= 4:
			return "Does it matter which door? You'll open the next one too.", true
		case c.WrongDoorsTotal >= 3:
			return "You keep opening the wrong ones.", true
		case b >= 3:
			return "You knew. You opened it anyway.", true
		default:
			return "Nothing behind that one.", true
		}
	case CueDecoy:
		switch {
		case b >= 4:
			return "You reach for every way out. None of them are the way out.", true
		case b >= 3:
			return "So close to the way out. But not quite.", true
		default:
			return "Not every door leads onward.", true
		}
	case CueKill:
		switch {
		case b >= 4:
			return "More of them. There are always more.", true
		case b >= 3:
			return "They keep coming. You keep firing.", true
		default:
			return "Efficient.", true
		}
	case CueItem:
		switch {
		case b >= 4:
			return "You still pick things up. Habits outlast their reasons.", true
		case b >= 3:
			return "Gathering things. As if it changes anything.", true
		default:
			return "You take what you find.", true
		}
	case CueSecret:
		switch {
		case b >= 4:
			return "A hidden room. As if finding it changes where you are.", true
		case b >= 3:
			return "You found the seam in things.", true
		default:
			return "A hidden place. Noted.", true
		}
	case CueDeath:
		switch {
		case b >= 4:
			return "It doesn't even stop you anymore.", true
		case c.Deaths >= 3:
			return "You've done this so many times now.", true
		case b >= 3:
			return "Death isn't a door either.", true
		default:
			return "Back again.", true
		}
	}
	return "", false
}
