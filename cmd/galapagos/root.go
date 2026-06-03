package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is the build version, injected at release time via -ldflags
// "-X main.version=...". It defaults to "dev" for local builds.
var version = "dev"

// rootCmd is the base galapagos command.
var rootCmd = &cobra.Command{
	Use:           "galapagos",
	Short:         "Visualize learning algorithms in real time",
	Long:          "Galapagos is a pluggable framework for visualizing learning algorithms. Its flagship demo evolves a population of cars to drive a procedurally generated track.",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
