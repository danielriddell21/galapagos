package main

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"time"

	"github.com/danielriddell21/galapagos/internal/agents/efficientcube"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/cube"
	rubix "github.com/danielriddell21/rubix/pkg/cube"
	"github.com/spf13/cobra"
)

var cubeKeymap = []string{"space pause", "+/- speed", "r new scrambles"}

// coupleScramble raises the training depth and search horizon to cover a
// scramble of the given target depth, never shrinking values the user set
// higher. The horizon keeps a small margin above the scramble so the solver can
// still find a slightly suboptimal solution.
func coupleScramble(target, scrambleK, maxDepth int) (k, depth int) {
	if scrambleK < target {
		scrambleK = target
	}
	if maxDepth < target+2 {
		maxDepth = target + 2
	}
	return scrambleK, maxDepth
}

func init() {
	var (
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
	)
	cmd := &cobra.Command{
		Use:   "cube",
		Short: "Learn to solve a Rubik's cube with EfficientCube (self-supervised policy + beam search)",
		Long:  "Trains an EfficientCube policy by self-supervision (predicting the move that reverses each scramble step), then solves scrambles with beam search. The window shows a cube being solved.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			seed = resolveSeed(cmd, 0, log)

			// Couple training depth and search horizon to the deepest scramble we
			// will face, so a deeper --scramble is always trained for and solvable
			// rather than silently undertrained. Beyond ~8 it stays best-effort.
			target := guiDepth
			if headless {
				target = evalDep
			}
			scrK, maxDepth = coupleScramble(target, scrK, maxDepth)
			log.Info("scramble", "depth", target, "train_k", scrK, "max_depth", maxDepth)

			policy, err := loadOrTrainPolicy(model, train, efficientcube.TrainConfig{
				Hidden: hidden, K: scrK, Batch: batch, Iters: iters, LR: lr, Seed: seed,
			}, log)
			if err != nil {
				return err
			}
			beam := efficientcube.BeamConfig{Width: beamW, MaxDepth: maxDepth}

			if headless {
				reportCubeEval(policy, evalDep, evalN, beam, seed)
				return nil
			}
			return cubeWall(policy, beam, cubes, guiDepth, seed, log)
		},
	}
	cmd.Flags().IntVar(&iters, "iters", 3000, "training iterations")
	cmd.Flags().IntVar(&batch, "batch", 256, "examples per training step")
	cmd.Flags().IntVar(&scrK, "scramble-k", 8, "max scramble length used for training (auto-raised to the scramble depth)")
	cmd.Flags().IntSliceVar(&hidden, "hidden", []int{128, 64}, "hidden layer sizes")
	cmd.Flags().Float64Var(&lr, "lr", 1e-3, "Adam learning rate")
	cmd.Flags().StringVar(&model, "model", "", "policy file to load (skip training) or save to")
	cmd.Flags().BoolVar(&train, "train", true, "train a policy (false loads --model)")
	cmd.Flags().IntVar(&evalN, "eval-scrambles", 100, "scrambles to solve in headless eval")
	cmd.Flags().IntVar(&evalDep, "eval-depth", 6, "scramble depth for headless eval")
	cmd.Flags().IntVar(&beamW, "beam-width", 2000, "beam search width (wider solves deeper scrambles more reliably)")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 20, "maximum solution length searched")
	cmd.Flags().IntVar(&guiDepth, "scramble", 8, "scramble depth shown in the window")
	cmd.Flags().IntVar(&cubes, "cubes", 1, "number of cubes solved in parallel in the window")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&headless, "headless", false, "train and evaluate without a window")
	rootCmd.AddCommand(cmd)
}

// loadOrTrainPolicy loads a saved policy when requested, otherwise trains one and
// optionally saves it.
func loadOrTrainPolicy(model string, train bool, cfg efficientcube.TrainConfig, log *slog.Logger) (*efficientcube.Policy, error) {
	if model != "" && !train {
		log.Info("loading policy", "path", model)
		return efficientcube.LoadPolicy(model)
	}
	log.Info("training", "iters", cfg.Iters, "scramble_k", cfg.K, "hidden", cfg.Hidden, "seed", cfg.Seed)
	p := efficientcube.Train(cfg, func(it int, loss, acc float64) {
		log.Info("train", "iter", it, "loss", loss, "acc", acc)
	})
	if model != "" {
		if err := p.Save(model); err != nil {
			return nil, err
		}
		log.Info("saved policy", "path", model)
	}
	return p, nil
}

// reportCubeEval beam-solves N scrambles in parallel and prints aggregate stats.
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

// cubeWall opens a window showing a grid of cubes being solved in parallel: the
// solutions are computed concurrently with SolveBatch, then animated together.
func cubeWall(p *efficientcube.Policy, beam efficientcube.BeamConfig, count, depth int, seed int64, log *slog.Logger) error {
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

	solvedCount := func() int {
		k := 0
		for i := range count {
			if states[i].IsSolved() {
				k++
			}
		}
		return k
	}

	run := guiRun{
		title:  "Galapagos — cube (efficientcube)",
		keymap: cubeKeymap,
		step: func() bool {
			pending := false
			for i := range count {
				if !states[i].IsSolved() && idx[i] < len(plans[i]) {
					pending = true
				}
			}
			if pending {
				// Advance the shared turn animation; apply the moves when it completes.
				if frac += 1.0 / turnFrames; frac >= 1 {
					frac = 0
					for i := range count {
						if !states[i].IsSolved() && idx[i] < len(plans[i]) {
							states[i] = states[i].Applied(plans[i][idx[i]])
							idx[i]++
						}
					}
				}
				return false
			}
			// All cubes finished; hold the solved cubes briefly before the next batch.
			hold++
			return hold >= 45
		},
		next: func() { generate(cur + int64(count)) },
		render: func(r core.Renderer) {
			for i := range count {
				ox := float64(i%cols) * (cube.NetW + gap)
				oy := float64(i/cols) * (cube.NetH + gap)
				turn, f := rubix.Move(0), 0.0
				if !states[i].IsSolved() && idx[i] < len(plans[i]) {
					turn, f = plans[i][idx[i]], frac
				}
				cube.RenderCube(r, states[i], turn, f, ox, oy)
			}
		},
		hud: func() []string {
			return []string{
				fmt.Sprintf("cubes %d  scramble %d", count, depth),
				fmt.Sprintf("solved %d/%d", solvedCount(), count),
			}
		},
		series: func() []float64 { return nil },
		bounds: func() (float64, float64, float64, float64, bool) {
			return 0, 0, float64(cols)*(cube.NetW+gap) - gap, float64(rows)*(cube.NetH+gap) - gap, true
		},
		leader:     func() (float64, float64, bool) { return 0, 0, false },
		sensors:    func() (core.Vec2, []core.Vec2, bool) { return core.Vec2{}, nil, false },
		regenerate: generate,
	}
	return launchGUI(run, log)
}
