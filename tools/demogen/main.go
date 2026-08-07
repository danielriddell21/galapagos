// Command demogen renders galapagos's documentation media headlessly: one
// short, deterministic clip per environment. It wires each environment exactly
// as its command does and draws through the software renderer — the same
// frames the window shows — so it needs no display.
//
// Regenerate every asset under docs/demos with:
//
//	just demos      // or: go run ./tools/demogen
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/galapagos/internal/cli"
)

const outDir = "docs/demos"

// seed is shared by every clip: one seed keeps the whole documentation set
// reproducible, and each environment draws its own stream from it.
const seed = 7

// fps and scale are the playback rate and downsampling every clip records at.
const (
	fps   = 25
	scale = 2
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demogen:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	for _, c := range clips() {
		if err := c.record(log); err != nil {
			return fmt.Errorf("%s: %w", c.env, err)
		}
	}
	return nil
}

// clip is one recorded run: which environment and agent, and how long the
// media plays for.
type clip struct {
	env string
	// ext is the output extension; an .mp4 records video instead of a GIF.
	ext      string
	agent    string
	opponent string
	frames   int
}

// clips is the documentation set: one clip per environment, each long enough
// to show the behaviour it is there to demonstrate. The learning environments
// train before they record, so cube and chess take by far the longest.
func clips() []clip {
	return []clip{
		{env: "race", ext: ".gif", frames: 220},
		{env: "cartpole", ext: ".gif", agent: "ga", frames: 180},
		{env: "maze", ext: ".gif", agent: "qlearning", frames: 180},
		{env: "cube", ext: ".gif", frames: 130},
		{env: "flappy", ext: ".gif", agent: "ga", frames: 130},
		{env: "chess", ext: ".gif", agent: "ga", opponent: "coevolution", frames: 95},
	}
}

func (c clip) record(log *slog.Logger) error {
	spec := cli.DemoSpec{
		Env:      c.env,
		Agent:    c.agent,
		Opponent: c.opponent,
		Seed:     seed,
		Rec: record.Options{
			Path:   filepath.Join(outDir, c.env+c.ext),
			Frames: c.frames,
			FPS:    fps,
			Scale:  scale,
		},
	}
	if err := cli.Demo(spec, log); err != nil {
		return fmt.Errorf("capture clip: %w", err)
	}
	return nil
}
