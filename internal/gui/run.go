package gui

import (
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/phrasing"
	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/status"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

func Run(cfg Config) error {
	seed := cfg.Seed
	if seed == 0 {
		seed = int64(rand.Uint64() >> 1)
	}
	fmt.Printf("pandemonium — seed %d\n", seed)

	settings := LoadSettings()

	// Carry the player's history forward: the commentary resumes near the deepest
	// level earlier runs reached, and the keeper folds this run back into the file.
	records := LoadRecords()
	keeper := NewRecordKeeper(records)

	overlay := hud.New()
	src, releaseNotice := noticeSource()
	defer releaseNotice()
	reporter := status.New(overlay, src, status.WithReach(records.Reach()))
	bus := telemetry.NewBus(reporter, keeper)

	// Sound is on by default but optional via the settings file (turn it off on a
	// headless machine with no audio device). If the engine fails to initialise we
	// also fall back to silence.
	var aud *Audio
	if settings.Sound {
		a, err := NewAudio()
		if err != nil {
			fmt.Fprintln(os.Stderr, "audio disabled:", err)
		} else {
			aud = a
		}
	}
	sess := newSession(seed, cfg.Width, cfg.Height, bus, aud)

	renderer := render.NewRenderer(render.DefaultConfig(), render.WithOverlay(overlay))
	game := New(sess.start, sess.next, renderer,
		WithOverlay(overlay),
		WithAudio(aud),
		WithSettings(settings),
		WithAttract(sess.attract),
		WithRecords(keeper),
		WithDifficulty(sess.setSkill),
	)
	if err := game.Run(); err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	return nil
}

func noticeSource() (status.Source, func()) {
	v, err := phrasing.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "notice phrasing disabled:", err)
		return status.NewTableSource(), func() {} // fall back to the scripted lines
	}
	return v, func() { _ = v.Close() }
}
