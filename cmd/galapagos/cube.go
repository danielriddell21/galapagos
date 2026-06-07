package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/danielriddell21/galapagos/internal/agents/efficientcube"
	"github.com/danielriddell21/galapagos/internal/envs/cube"
	rubix "github.com/danielriddell21/rubix/pkg/cube"
	"github.com/spf13/cobra"
)

var cubeKeymap = []string{"space pause", "+/- speed", "r new scramble"}

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
		maxSteps int
		seed     int64
		headless bool
	)
	cmd := &cobra.Command{
		Use:   "cube",
		Short: "Learn to solve a Rubik's cube with EfficientCube (self-supervised policy + beam search)",
		Long:  "Trains an EfficientCube policy by self-supervision (predicting the move that reverses each scramble step), then solves scrambles with beam search. The window shows the learned policy turning the cube toward solved.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			seed = resolveSeed(cmd, 0, log)

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
			return cubeGUI(policy, beam, guiDepth, maxSteps, seed, log)
		},
	}
	cmd.Flags().IntVar(&iters, "iters", 2000, "training iterations")
	cmd.Flags().IntVar(&batch, "batch", 256, "examples per training step")
	cmd.Flags().IntVar(&scrK, "scramble-k", 6, "max scramble length used for training")
	cmd.Flags().IntSliceVar(&hidden, "hidden", []int{128, 64}, "hidden layer sizes")
	cmd.Flags().Float64Var(&lr, "lr", 1e-3, "Adam learning rate")
	cmd.Flags().StringVar(&model, "model", "", "policy file to load (skip training) or save to")
	cmd.Flags().BoolVar(&train, "train", true, "train a policy (false loads --model)")
	cmd.Flags().IntVar(&evalN, "eval-scrambles", 100, "scrambles to solve in headless eval")
	cmd.Flags().IntVar(&evalDep, "eval-depth", 6, "scramble depth for headless eval")
	cmd.Flags().IntVar(&beamW, "beam-width", 200, "beam search width")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 18, "maximum solution length searched")
	cmd.Flags().IntVar(&guiDepth, "scramble", 6, "scramble depth shown in the window")
	cmd.Flags().IntVar(&maxSteps, "max-steps", 40, "maximum moves per episode in the window")
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

// reportCubeEval beam-solves N scrambles at the given depth and prints aggregate
// statistics.
func reportCubeEval(p *efficientcube.Policy, depth, n int, beam efficientcube.BeamConfig, seed int64) {
	start := time.Now()
	solved, totLen, totNodes := 0, 0, 0
	for i := range n {
		c := rubix.ScrambledCube(depth, seed+int64(i))
		res := p.Solve(c, beam)
		if res.Solved {
			solved++
			totLen += len(res.Moves)
		}
		totNodes += res.Nodes
	}
	avgLen := 0.0
	if solved > 0 {
		avgLen = float64(totLen) / float64(solved)
	}
	fmt.Printf("depth %d: solved %d/%d (%.0f%%), avg length %.1f, avg nodes %d, %s\n",
		depth, solved, n, 100*float64(solved)/float64(n), avgLen, totNodes/n, time.Since(start).Round(time.Millisecond))
}

// cubeGUI opens the window where the beam-search solution turns the cube to
// solved, one move per frame.
func cubeGUI(p *efficientcube.Policy, beam efficientcube.BeamConfig, depth, maxSteps int, seed int64, log *slog.Logger) error {
	env := cube.New(cube.Config{ScrambleDepth: depth, MaxSteps: maxSteps})
	agent := efficientcube.NewAgent(p, beam)
	caps := runCaps{
		keymap: cubeKeymap,
		bounds: boundsOf(env),
		extraHUD: func() []string {
			return []string{
				fmt.Sprintf("beam width %d", beam.Width),
				fmt.Sprintf("solution %d/%d", agent.Move(), agent.SolutionLen()),
			}
		},
	}
	return launchGUI(onlineGUI("Galapagos — cube (efficientcube)", env, agent, maxSteps, seed, caps), log)
}
