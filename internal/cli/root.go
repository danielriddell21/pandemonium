package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/pandemonium/internal/gui"
)

func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

func newRootCmd(version string) *cobra.Command {
	var cfg gui.Config

	cmd := &cobra.Command{
		Use:           "pandemonium",
		Short:         "A procedurally generated DOOM-style raycaster.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := gui.Run(cfg); err != nil {
				return fmt.Errorf("run gui: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().Int64Var(&cfg.Seed, "seed", 0, "world seed (0 picks a random seed)")
	cmd.Flags().IntVar(&cfg.Width, "width", 48, "level width in tiles")
	cmd.Flags().IntVar(&cfg.Height, "height", 32, "level height in tiles")

	cmd.AddCommand(completionCmd())
	return cmd
}
