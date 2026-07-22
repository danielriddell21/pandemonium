package phrasing

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/danielriddell21/narrata"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/status"
)

//go:embed personas.json
var personasJSON []byte

const persona = "house"

const genTimeout = 2 * time.Second

type Source struct {
	engine   *narrata.Engine
	scripted status.Source
	tmp      string
}

func New() (*Source, error) {
	tmp, err := writeTemp(personasJSON)
	if err != nil {
		return nil, err
	}
	engine, err := narrata.New(narrata.Config{
		PersonasPath:   tmp,
		DefaultPersona: persona,
		Text:           narrata.TextConfig{Backend: "native"},
		MaxConcurrent:  2,
		Timeout:        genTimeout,
	})
	if err != nil {
		_ = os.Remove(tmp)
		return nil, fmt.Errorf("start phrasing engine: %w", err)
	}
	return &Source{
		engine:   engine,
		scripted: status.NewTableSource(),
		tmp:      tmp,
	}, nil
}

func (s *Source) Request(cue status.Cue, emit func(status.Line)) {
	var line status.Line
	var got bool
	s.scripted.Request(cue, func(l status.Line) { line, got = l, true })
	if !got {
		return // the scripted policy stayed silent for this cue
	}
	if line.Channel != hud.Notice {
		emit(line) // diagnostics are never rewritten
		return
	}
	if text, ok := s.generate(context.Background(), cue); ok {
		line.Text = text
	}
	emit(line) // the scripted wording stands in when nothing was generated
}

func (s *Source) generate(ctx context.Context, cue status.Cue) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, genTimeout)
	defer cancel()
	res, err := s.engine.Generate(ctx, narrata.Request{
		PersonaID:   persona,
		Event:       eventFor(cue),
		Data:        dataFor(cue),
		Output:      narrata.OutputText,
		Constraints: narrata.Constraints{MaxWords: 14},
	})
	if err != nil || res.Silent || res.Text == "" {
		return "", false
	}
	return res.Text, true
}

func (s *Source) Close() error {
	if s.tmp != "" {
		_ = os.Remove(s.tmp)
	}
	if err := s.engine.Close(); err != nil {
		return fmt.Errorf("close phrasing engine: %w", err)
	}
	return nil
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

func writeTemp(b []byte) (string, error) {
	f, err := os.CreateTemp("", "pandemonium-personas-*.json")
	if err != nil {
		return "", fmt.Errorf("create personas file: %w", err)
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("write personas file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("close personas file: %w", err)
	}
	return f.Name(), nil
}
