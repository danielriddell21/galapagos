package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	recordPath   string
	recordFrames int
	recordFPS    int
	recordScale  int
)

var rootCmd = &cobra.Command{
	Use:           "galapagos",
	Short:         "Visualize learning algorithms in real time",
	Long:          "Galapagos is a pluggable framework for visualizing learning algorithms. Its flagship demo evolves a population of cars to drive a procedurally generated track.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	f := rootCmd.PersistentFlags()
	f.StringVar(&recordPath, "record", "", "record the window to an animated GIF at this path, then exit")
	f.IntVar(&recordFrames, "record-frames", 150, "number of frames to record")
	f.IntVar(&recordFPS, "record-fps", 25, "frames per second of the recording")
	f.IntVar(&recordScale, "record-scale", 2, "integer downscale factor for the GIF")
	rootCmd.AddCommand(completionCmd())
}

func Execute(version string) error {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}
