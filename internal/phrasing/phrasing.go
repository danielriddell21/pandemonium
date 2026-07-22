package phrasing

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/danielriddell21/crucible/narrate"

	"github.com/danielriddell21/pandemonium/internal/status"
)

//go:embed personas.json
var personasJSON []byte

const persona = "house"

const genTimeout = 2 * time.Second

// New returns a status source that rewrites the scripted notice lines
// through the narrata "house" persona, standing the scripted wording back in
// whenever generation declines, fails, or times out.
func New() (*narrate.Source[status.Cue], error) {
	src, err := narrate.New(narrate.Config{
		Personas:      personasJSON,
		Persona:       persona,
		MaxWords:      14,
		Timeout:       genTimeout,
		MaxConcurrent: 2,
	}, status.NewTableSource(), eventFor, dataFor)
	if err != nil {
		return nil, fmt.Errorf("start phrasing engine: %w", err)
	}
	return src, nil
}

func eventFor(c status.Cue) string {
	if c.Arrival {
		return "arrival"
	}
	switch c.Kind {
	case status.CueExit:
		return "level_complete"
	case status.CueKill:
		return "enemy_killed"
	case status.CueItem:
		return "item_taken"
	case status.CueSecret:
		return "secret_found"
	case status.CueDeath:
		return "player_died"
	case status.CueWrongDoor:
		return "wrong_turn"
	case status.CueDecoy:
		return "decoy"
	case status.CueFork:
		return "fork"
	default:
		return "event"
	}
}

func dataFor(c status.Cue) map[string]any {
	return map[string]any{
		"level":   c.Level + 1, // the {level} placeholder is 1-based for the player
		"cleared": c.LevelsCleared,
		"deaths":  c.Deaths,
		"rushing": c.Rushing,
	}
}
