package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/crucible/record"
)

var rec record.Options

var rootCmd = &cobra.Command{
	Use:           "galapagos",
	Short:         "Visualize learning algorithms in real time",
	Long:          "Galapagos is a pluggable framework for visualizing learning algorithms. Its flagship demo evolves a population of cars to drive a procedurally generated track.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Preserve galapagos's own defaults; the flag names come from crucible.
	rec.FPS, rec.Frames, rec.Scale = 25, 150, 2
	rec.AddFlags(rootCmd.PersistentFlags())
	rootCmd.AddCommand(completionCmd())
}

func Execute(version string) error {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}
