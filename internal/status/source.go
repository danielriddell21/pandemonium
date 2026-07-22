package status

import (
	"github.com/danielriddell21/crucible/hud"
	cstatus "github.com/danielriddell21/crucible/status"
)

type CueKind uint8

const (
	CueExit CueKind = iota

	CueWrongDoor

	CueDecoy

	CueFork

	CueKill

	CueItem

	CueSecret

	CueDeath
)

type Cue struct {
	Kind          CueKind
	Level         int
	OptimalChoice bool

	LevelsCleared   int
	Deaths          int
	WrongDoorsTotal int
	ExploreScore    float64
	Rushing         bool

	Arrival bool
}

// Line is one status message ready for the overlay, shared with the family
// through crucible/status.
type Line = cstatus.Line

const messageFrames = 150

const deepestBand = 4

type tableSource struct{}

func NewTableSource() cstatus.Source[Cue] { return tableSource{} }

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

func band1Whisper(c Cue) (string, bool) {
	switch c.Kind {
	case CueDeath:
		return "Hm. Again.", true
	case CueSecret:
		return "Something tucked away.", true
	}
	return "", false
}

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

func noticeText(c Cue, b int) (string, bool) {
	switch c.Kind {
	case CueExit:
		return exitNotice(c, b)
	case CueWrongDoor:
		return wrongDoorNotice(c, b)
	case CueDecoy:
		return decoyNotice(b)
	case CueKill:
		return killNotice(b)
	case CueItem:
		return itemNotice(b)
	case CueSecret:
		return secretNotice(b)
	case CueDeath:
		return deathNotice(c, b)
	}
	return "", false
}

func exitNotice(c Cue, b int) (string, bool) {
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
}

func wrongDoorNotice(c Cue, b int) (string, bool) {
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
}

func decoyNotice(b int) (string, bool) {
	switch {
	case b >= 4:
		return "You reach for every way out. None of them are the way out.", true
	case b >= 3:
		return "So close to the way out. But not quite.", true
	default:
		return "Not every door leads onward.", true
	}
}

func killNotice(b int) (string, bool) {
	switch {
	case b >= 4:
		return "More of them. There are always more.", true
	case b >= 3:
		return "They keep coming. You keep firing.", true
	default:
		return "Efficient.", true
	}
}

func itemNotice(b int) (string, bool) {
	switch {
	case b >= 4:
		return "You still pick things up. Habits outlast their reasons.", true
	case b >= 3:
		return "Gathering things. As if it changes anything.", true
	default:
		return "You take what you find.", true
	}
}

func secretNotice(b int) (string, bool) {
	switch {
	case b >= 4:
		return "A hidden room. As if finding it changes where you are.", true
	case b >= 3:
		return "You found the seam in things.", true
	default:
		return "A hidden place. Noted.", true
	}
}

func deathNotice(c Cue, b int) (string, bool) {
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
