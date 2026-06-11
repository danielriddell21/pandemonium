// Package hud holds transient heads-up display state shared between the game loop
// and the renderer: a single message line that shows for a set number of frames
// and then clears. It is plain state with no dependencies, so any layer can read
// or write it without coupling.
package hud

// Channel selects how a message is presented and who sees it.
type Channel uint8

const (
	// Diagnostic messages are development/playtest readouts. They are only drawn
	// in debug builds.
	Diagnostic Channel = iota
	// Notice messages are player-facing in-game notices and are always drawn.
	Notice
)

// Overlay is a one-line message with a frame-countdown lifetime. It is written by
// whatever produces messages, advanced once per frame by the game loop, and read
// by the renderer. It is used from a single goroutine and needs no locking.
type Overlay struct {
	text      string
	channel   Channel
	remaining int
}

// New returns an empty overlay.
func New() *Overlay {
	return &Overlay{}
}

// Post shows text on the given channel for the given number of frames, replacing
// any current message. Empty text or a non-positive frame count clears it.
func (o *Overlay) Post(text string, frames int, ch Channel) {
	if text == "" || frames <= 0 {
		o.text, o.channel, o.remaining = "", Diagnostic, 0
		return
	}
	o.text, o.channel, o.remaining = text, ch, frames
}

// Tick advances the message lifetime by one frame, clearing it when it expires.
func (o *Overlay) Tick() {
	if o.remaining > 0 {
		o.remaining--
		if o.remaining == 0 {
			o.text = ""
		}
	}
}

// Active returns the current message, its channel, and whether one is showing.
func (o *Overlay) Active() (string, Channel, bool) {
	return o.text, o.channel, o.remaining > 0
}
