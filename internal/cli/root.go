package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "galapagos",
	Short:         "Visualize learning algorithms in real time",
	Long:          "Galapagos is a pluggable framework for visualizing learning algorithms. Its flagship demo evolves a population of cars to drive a procedurally generated track.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(completionCmd())
}

func Execute(version string) error {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}
