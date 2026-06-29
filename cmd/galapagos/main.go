// Command galapagos runs the Galapagos learning-visualization framework. The
// flagship subcommand, race, evolves a population of cars on a procedural track.
package main

import (
	"fmt"
	"os"

	"github.com/danielriddell21/galapagos/internal/cli"
)

// version is the build version, injected at release time via -ldflags
// "-X main.version=...". It defaults to "dev" for local builds.
var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
