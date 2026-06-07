package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is the build version, injected at release time via -ldflags
// "-X main.version=...". It defaults to "dev" for local builds.
var version = "dev"

// Recording flags. When recordPath is set, a windowed run captures frames from
// the Ebiten screen and writes an animated GIF, then exits. They are read by the
// Ebiten game (the headless paths and the non-GUI build ignore them).
var (
	recordPath   string
	recordFrames int
	recordFPS    int
	recordScale  int
)

// rootCmd is the base galapagos command.
var rootCmd = &cobra.Command{
	Use:           "galapagos",
	Short:         "Visualize learning algorithms in real time",
	Long:          "Galapagos is a pluggable framework for visualizing learning algorithms. Its flagship demo evolves a population of cars to drive a procedurally generated track.",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	f := rootCmd.PersistentFlags()
	f.StringVar(&recordPath, "record", "", "record the window to an animated GIF at this path, then exit")
	f.IntVar(&recordFrames, "record-frames", 150, "number of frames to record")
	f.IntVar(&recordFPS, "record-fps", 25, "frames per second of the recording")
	f.IntVar(&recordScale, "record-scale", 2, "integer downscale factor for the GIF")
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
