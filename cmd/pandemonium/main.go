// Command pandemonium is a procedurally generated, DOOM-style raycaster. Each
// level is generated from a seed; reaching the exit generates the next.
package main

import (
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/pandemonium/internal/app"
	"github.com/danielriddell21/pandemonium/internal/hud"
	"github.com/danielriddell21/pandemonium/internal/render"
	"github.com/danielriddell21/pandemonium/internal/status"
	"github.com/danielriddell21/pandemonium/internal/telemetry"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var (
		seed   int64
		width  int
		height int
	)

	cmd := &cobra.Command{
		Use:           "pandemonium",
		Short:         "A procedurally generated DOOM-style raycaster.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(seed, width, height)
		},
	}

	cmd.Flags().Int64Var(&seed, "seed", 0, "world seed (0 picks a random seed)")
	cmd.Flags().IntVar(&width, "width", 48, "level width in tiles")
	cmd.Flags().IntVar(&height, "height", 32, "level height in tiles")
	return cmd
}

// run sets up the telemetry bus, the level session, the HUD overlay and the
// renderer, then opens the window and plays.
func run(seed int64, width, height int) error {
	if seed == 0 {
		seed = int64(rand.Uint64() >> 1)
	}
	fmt.Printf("pandemonium — seed %d\n", seed)

	overlay := hud.New()
	bus := telemetry.NewBus(status.New(overlay, status.NewTableSource()))

	settings := app.LoadSettings()

	// Sound is on by default but optional via the settings file (turn it off on a
	// headless machine with no audio device). If the engine fails to initialise we
	// also fall back to silence.
	var aud *app.Audio
	if settings.Sound {
		a, err := app.NewAudio()
		if err != nil {
			fmt.Fprintln(os.Stderr, "audio disabled:", err)
		} else {
			aud = a
		}
	}
	sess := newSession(seed, width, height, bus, aud)

	renderer := render.NewRenderer(render.DefaultConfig(), render.WithOverlay(overlay))
	game := app.New(sess.start(), renderer, sess.next,
		app.WithOverlay(overlay),
		app.WithAudio(aud),
		app.WithSettings(settings),
		app.WithAttract(sess.attract),
	)
	return game.Run()
}
