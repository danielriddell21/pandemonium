package gui

import (
	"testing"

	"github.com/danielriddell21/pandemonium/internal/status"
)

func TestNoticeSource(t *testing.T) {
	src, done := noticeSource()
	if src == nil {
		t.Fatal("expected a status source")
	}
	defer done()

	var emitted bool
	src.Request(status.Cue{Kind: status.CueExit, Level: 6, LevelsCleared: 6}, func(status.Line) { emitted = true })
	if !emitted {
		t.Error("expected the source to emit a line for a notice cue")
	}
}
