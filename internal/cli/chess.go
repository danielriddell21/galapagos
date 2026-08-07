package cli

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/galapagos/internal/agents/evochess"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/chess"
	"github.com/danielriddell21/galapagos/internal/gui"
	"github.com/danielriddell21/galapagos/internal/sim"
)

type rawAction []float64

func (a rawAction) Vector() []float64 { return a }

type policyAgent struct{ p chess.Policy }

func (a policyAgent) Act(s core.State) core.Action                                 { return a.p(s) }
func (policyAgent) Observe(core.State, core.Action, core.Reward, core.State, bool) {}
func (policyAgent) EndEpisode(core.Reward)                                         {}

type bestPolicyProvider interface {
	BestPolicy() func(obs []float64) []float64
}

type chessPolicyProvider interface {
	FrozenPolicy() func(core.State) core.Action
}

func bestChessPolicy(pop core.PopulationAgent) chess.Policy {
	switch p := pop.(type) {
	case chessPolicyProvider:
		return p.FrozenPolicy()
	case bestPolicyProvider:
		fwd := p.BestPolicy()
		return func(s core.State) core.Action { return rawAction(fwd(s.Observation())) }
	default:
		return nil
	}
}

func newChessAgent(agent string, obs, act core.Spec, population, depth int, seed int64) (core.PopulationAgent, error) {
	if agent == "evochess" {
		c := evochess.DefaultConfig()
		c.Population, c.Depth, c.Seed = population, depth, seed
		return evochess.New(c), nil
	}
	return newPopulationAgent(agent, obs, act, population, seed)
}

// chessParams are the chess command's tunables. The values defaultChess
// returns are the flag defaults, and the demo clip runs with them unchanged,
// so the recorded media shows what the command does.
type chessParams struct {
	agent       string
	opponent    string
	generations int
	population  int
	maxPlies    int
	depth       int
	seed        int64
	headless    bool
}

func defaultChess() chessParams {
	return chessParams{agent: "ga", opponent: "coevolution", generations: 20, population: 40, maxPlies: 60, depth: 2}
}

func init() {
	c := defaultChess()
	var seed int64
	cmd := &cobra.Command{
		Use:   "chess",
		Short: "Evolve a chess player (ga, neat, or evochess) against a co-evolving rival",
		Long: "Evolves a population to play chess (wrapping the gambit engine). ga and neat " +
			"emit a move from a network; evochess evolves a board-evaluation function and " +
			"plays by alpha-beta search. With --opponent coevolution each generation is scored " +
			"against a frozen copy of the previous generation's best (a hall-of-fame champion " +
			"that strengthens over time); --opponent random plays a uniform-random mover. The " +
			"windowed mode trains, then shows the evolved player (White) against a random opponent.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			c.seed = resolveSeed(cmd, 0, log)

			pop, best, err := evolveChess(c, log)
			if err != nil {
				return err
			}
			if c.headless {
				fmt.Printf("best fitness %.3f\n", best)
				return nil
			}
			return watchChess(pop, c.maxPlies, c.seed, c.agent, log)
		},
	}
	cmd.Flags().StringVar(&c.agent, "agent", c.agent, "agent: ga, neat, or evochess")
	cmd.Flags().StringVar(&c.opponent, "opponent", c.opponent, "opponent: coevolution or random")
	cmd.Flags().IntVar(&c.generations, "generations", c.generations, "generations to evolve")
	cmd.Flags().IntVar(&c.population, "population", c.population, "population size")
	cmd.Flags().IntVar(&c.maxPlies, "max-plies", c.maxPlies, "maximum plies before a game is drawn")
	cmd.Flags().IntVar(&c.depth, "depth", c.depth, "alpha-beta search depth (evochess only)")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&c.headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}

// validate rejects agent and opponent names the chess environment cannot wire up.
func (c chessParams) validate() error {
	if c.agent != "ga" && c.agent != "neat" && c.agent != "evochess" {
		return fmt.Errorf("chess supports only the ga, neat, and evochess agents")
	}
	if c.opponent != "random" && c.opponent != "coevolution" {
		return fmt.Errorf("unknown opponent %q (want random or coevolution)", c.opponent)
	}
	return nil
}

func evolveChess(c chessParams, log *slog.Logger) (core.PopulationAgent, float64, error) {
	if err := c.validate(); err != nil {
		return nil, 0, err
	}
	maxPlies, seed := c.maxPlies, c.seed
	coevolve := c.opponent == "coevolution"
	sample := chess.New(chess.Config{MaxPlies: maxPlies})
	pop, err := newChessAgent(c.agent, sample.ObservationSpec(), sample.ActionSpec(), c.population, c.depth, seed)
	if err != nil {
		return nil, 0, err
	}
	tel := sim.NewTelemetry(c.generations, log)
	log.Info("evolving", "agent", c.agent, "opponent", c.opponent,
		"generations", c.generations, "population", c.population, "seed", seed)

	var champion chess.Policy
	evaluate := func() []float64 {
		opp := champion // snapshot for this generation's factory
		factory := sim.EnvFactory(func() core.MultiEnvironment {
			return sim.AsMulti(func() core.Environment { return chess.New(chess.Config{MaxPlies: maxPlies, Opponent: opp}) })
		})
		return sim.EvaluateParallel(factory, individuals(pop), maxPlies, seed)
	}

	var fitness []float64
	for range c.generations {
		fitness = evaluate()
		tel.Publish(sim.StatsFrom(pop.Generation(), fitness))
		if coevolve {
			champion = bestChessPolicy(pop)
		}
		pop.Evolve()
	}
	// Evaluate the final (just-evolved) generation so its best is meaningful.
	fitness = evaluate()
	return pop, slices.Max(fitness), nil
}

func watchChess(pop core.PopulationAgent, maxPlies int, seed int64, agent string, log *slog.Logger) error {
	cfg, err := chessConfig(pop, maxPlies, seed, agent)
	if err != nil {
		return err
	}
	if err := gui.Run(cfg, log); err != nil {
		return fmt.Errorf("run gui: %w", err)
	}
	return nil
}

// chessConfig sets the evolved player (White) against a random opponent and
// wires the game into a run the display can drive. It is display-free, so the
// same wiring backs the window and the headless recordings in tools/demogen.
func chessConfig(pop core.PopulationAgent, maxPlies int, seed int64, agent string) (gui.Config, error) {
	best := bestChessPolicy(pop)
	if best == nil {
		return gui.Config{}, fmt.Errorf("agent %q cannot expose a best policy", agent)
	}
	env := chess.New(chess.Config{MaxPlies: maxPlies}) // random Black opponent
	caps := runCaps{keymap: onlineKeymap, bounds: boundsOf(env), leader: noLeader, sensors: noSensors}
	return onlineGUI("Galapagos — chess ("+agent+")", env, policyAgent{best}, maxPlies, seed, caps), nil
}
