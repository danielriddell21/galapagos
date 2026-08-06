package cli

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"time"

	rubix "github.com/danielriddell21/rubix/pkg/cube"
	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/agents/efficientcube"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/cube"
	"github.com/danielriddell21/galapagos/internal/gui"
)

var cubeKeymap = []string{"space pause", "+/- speed", "r new scrambles"}

func coupleScramble(target, scrambleK, maxDepth int) (k, depth int) {
	if scrambleK < target {
		scrambleK = target
	}
	if maxDepth < target+2 {
		maxDepth = target + 2
	}
	return scrambleK, maxDepth
}

// cubeParams are the cube command's tunables. The values defaultCube returns
// are the flag defaults, and the demo clip runs with them unchanged, so the
// recorded media shows what the command does.
type cubeParams struct {
	// training
	iters  int
	batch  int
	scrK   int
	hidden []int
	lr     float64
	model  string
	train  bool
	// solving / eval
	evalN    int
	evalDep  int
	beamW    int
	maxDepth int
	// run
	guiDepth int
	cubes    int
	seed     int64
	headless bool
}

func defaultCube() cubeParams {
	return cubeParams{
		iters: 3000, batch: 256, scrK: 8, hidden: []int{128, 64}, lr: 1e-3, train: true,
		evalN: 100, evalDep: 6, beamW: 2000, maxDepth: 20, guiDepth: 8, cubes: 1,
	}
}

// solve trains or loads the policy and returns it with the beam search the
// scramble depth calls for.
func (c cubeParams) solve(log *slog.Logger) (*efficientcube.Policy, efficientcube.BeamConfig, error) {
	// Couple training depth and search horizon to the deepest scramble we
	// will face, so a deeper --scramble is always trained for and solvable
	// rather than silently undertrained. Beyond ~8 it stays best-effort.
	target := c.guiDepth
	if c.headless {
		target = c.evalDep
	}
	scrK, maxDepth := coupleScramble(target, c.scrK, c.maxDepth)
	log.Info("scramble", "depth", target, "train_k", scrK, "max_depth", maxDepth)

	policy, err := loadOrTrainPolicy(c.model, c.train, efficientcube.TrainConfig{
		Hidden: c.hidden, K: scrK, Batch: c.batch, Iters: c.iters, LR: c.lr, Seed: c.seed,
	}, log)
	if err != nil {
		return nil, efficientcube.BeamConfig{}, err
	}
	return policy, efficientcube.BeamConfig{Width: c.beamW, MaxDepth: maxDepth}, nil
}

func init() {
	c := defaultCube()
	var seed int64
	cmd := &cobra.Command{
		Use:   "cube",
		Short: "Learn to solve a Rubik's cube with EfficientCube (self-supervised policy + beam search)",
		Long:  "Trains an EfficientCube policy by self-supervision (predicting the move that reverses each scramble step), then solves scrambles with beam search. The window shows a cube being solved.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			c.seed = resolveSeed(cmd, 0, log)

			policy, beam, err := c.solve(log)
			if err != nil {
				return err
			}
			if c.headless {
				reportCubeEval(policy, c.evalDep, c.evalN, beam, c.seed)
				return nil
			}
			return cubeWall(policy, beam, c.cubes, c.guiDepth, c.seed, log)
		},
	}
	cmd.Flags().IntVar(&c.iters, "iters", c.iters, "training iterations")
	cmd.Flags().IntVar(&c.batch, "batch", c.batch, "examples per training step")
	cmd.Flags().IntVar(&c.scrK, "scramble-k", c.scrK, "max scramble length used for training (auto-raised to the scramble depth)")
	cmd.Flags().IntSliceVar(&c.hidden, "hidden", c.hidden, "hidden layer sizes")
	cmd.Flags().Float64Var(&c.lr, "lr", c.lr, "Adam learning rate")
	cmd.Flags().StringVar(&c.model, "model", c.model, "policy file to load (skip training) or save to")
	cmd.Flags().BoolVar(&c.train, "train", c.train, "train a policy (false loads --model)")
	cmd.Flags().IntVar(&c.evalN, "eval-scrambles", c.evalN, "scrambles to solve in headless eval")
	cmd.Flags().IntVar(&c.evalDep, "eval-depth", c.evalDep, "scramble depth for headless eval")
	cmd.Flags().IntVar(&c.beamW, "beam-width", c.beamW, "beam search width (wider solves deeper scrambles more reliably)")
	cmd.Flags().IntVar(&c.maxDepth, "max-depth", c.maxDepth, "maximum solution length searched")
	cmd.Flags().IntVar(&c.guiDepth, "scramble", c.guiDepth, "scramble depth shown in the window")
	cmd.Flags().IntVar(&c.cubes, "cubes", c.cubes, "number of cubes solved in parallel in the window")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&c.headless, "headless", false, "train and evaluate without a window")
	rootCmd.AddCommand(cmd)
}

func loadOrTrainPolicy(model string, train bool, cfg efficientcube.TrainConfig, log *slog.Logger) (*efficientcube.Policy, error) {
	if model != "" && !train {
		log.Info("loading policy", "path", model)
		p, err := efficientcube.LoadPolicy(model)
		if err != nil {
			return nil, fmt.Errorf("load policy %q: %w", model, err)
		}
		return p, nil
	}
	log.Info("training", "iters", cfg.Iters, "scramble_k", cfg.K, "hidden", cfg.Hidden, "seed", cfg.Seed)
	p := efficientcube.Train(cfg, func(it int, loss, acc float64) {
		log.Info("train", "iter", it, "loss", loss, "acc", acc)
	})
	if model != "" {
		if err := p.Save(model); err != nil {
			return nil, fmt.Errorf("save policy %q: %w", model, err)
		}
		log.Info("saved policy", "path", model)
	}
	return p, nil
}

func reportCubeEval(p *efficientcube.Policy, depth, n int, beam efficientcube.BeamConfig, seed int64) {
	scrambles := make([]rubix.Cube, n)
	for i := range n {
		scrambles[i] = rubix.ScrambledCube(depth, seed+int64(i))
	}
	start := time.Now()
	results := p.SolveBatch(scrambles, beam)
	elapsed := time.Since(start)

	solved, totLen, totNodes := 0, 0, 0
	for _, r := range results {
		if r.Solved {
			solved++
			totLen += len(r.Moves)
		}
		totNodes += r.Nodes
	}
	avgLen := 0.0
	if solved > 0 {
		avgLen = float64(totLen) / float64(solved)
	}
	fmt.Printf("depth %d: solved %d/%d (%.0f%%), avg length %.1f, avg nodes %d, %s\n",
		depth, solved, n, 100*float64(solved)/float64(n), avgLen, totNodes/n, elapsed.Round(time.Millisecond))
}

func cubeWall(p *efficientcube.Policy, beam efficientcube.BeamConfig, count, depth int, seed int64, log *slog.Logger) error {
	if err := gui.Run(cubeConfig(p, beam, count, depth, seed), log); err != nil {
		return fmt.Errorf("run gui: %w", err)
	}
	return nil
}

// cubeConfig wires a wall of cubes and their solver into a run the display can
// drive. It is display-free, so the same wiring backs the window and the
// headless recordings in tools/demogen.
func cubeConfig(p *efficientcube.Policy, beam efficientcube.BeamConfig, count, depth int, seed int64) gui.Config {
	const gap = cube.NetW * 0.15
	cols := int(math.Ceil(math.Sqrt(float64(count))))
	rows := (count + cols - 1) / cols

	states := make([]rubix.Cube, count)
	plans := make([][]rubix.Move, count)
	idx := make([]int, count)
	cur := seed
	hold := 0
	frac := 0.0
	const turnFrames = 6 // frames to animate one move's rotation

	generate := func(s int64) {
		cur = s
		for i := range count {
			states[i] = rubix.ScrambledCube(depth, s+int64(i))
		}
		for i, res := range p.SolveBatch(states, beam) {
			plans[i] = res.Moves
			idx[i] = 0
		}
		hold = 0
		frac = 0
	}
	generate(seed)

	solvedCount := func() int { return countSolved(states) }

	run := gui.Config{Title: "Galapagos — cube (efficientcube)"}
	run.Keymap = cubeKeymap
	run.Step = func() bool {
		if !cubesPending(states, plans, idx) {
			// All cubes finished; hold the solved cubes briefly before the next batch.
			hold++
			return hold >= 45
		}
		// Advance the shared turn animation; apply the moves when it completes.
		if frac += 1.0 / turnFrames; frac >= 1 {
			frac = 0
			advanceCubes(states, plans, idx)
		}
		return false
	}
	run.Next = func() { generate(cur + int64(count)) }
	run.Render = func(r core.Renderer) {
		for i := range count {
			ox := float64(i%cols) * (cube.NetW + gap)
			oy := float64(i/cols) * (cube.NetH + gap)
			turn, f := rubix.Move(0), 0.0
			if !states[i].IsSolved() && idx[i] < len(plans[i]) {
				turn, f = plans[i][idx[i]], frac
			}
			cube.RenderCube(r, states[i], turn, f, ox, oy)
		}
	}
	run.HUD = func() []string {
		return []string{
			fmt.Sprintf("cubes %d  scramble %d", count, depth),
			fmt.Sprintf("solved %d/%d", solvedCount(), count),
		}
	}
	run.Series = func() []float64 { return nil }
	run.Bounds = func() (float64, float64, float64, float64, bool) {
		return 0, 0, float64(cols)*(cube.NetW+gap) - gap, float64(rows)*(cube.NetH+gap) - gap, true
	}
	run.Leader = func() (float64, float64, bool) { return 0, 0, false }
	run.Sensors = func() (core.Vec2, []core.Vec2, bool) { return core.Vec2{}, nil, false }
	run.Regenerate = generate
	return run
}

func cubesPending(states []rubix.Cube, plans [][]rubix.Move, idx []int) bool {
	for i := range states {
		if !states[i].IsSolved() && idx[i] < len(plans[i]) {
			return true
		}
	}
	return false
}

func advanceCubes(states []rubix.Cube, plans [][]rubix.Move, idx []int) {
	for i := range states {
		if !states[i].IsSolved() && idx[i] < len(plans[i]) {
			states[i] = states[i].Applied(plans[i][idx[i]])
			idx[i]++
		}
	}
}

func countSolved(states []rubix.Cube) int {
	k := 0
	for i := range states {
		if states[i].IsSolved() {
			k++
		}
	}
	return k
}
