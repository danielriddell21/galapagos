package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/danielriddell21/galapagos/internal/agents/evochess"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/envs/chess"
	"github.com/danielriddell21/galapagos/internal/sim"
	"github.com/spf13/cobra"
)

// rawAction carries a network's raw output vector as an action.
type rawAction []float64

func (a rawAction) Vector() []float64 { return a }

// policyAgent plays a fixed chess policy (no learning), for the exhibition game.
type policyAgent struct{ p chess.Policy }

func (a policyAgent) Act(s core.State) core.Action                                 { return a.p(s) }
func (policyAgent) Observe(core.State, core.Action, core.Reward, core.State, bool) {}
func (policyAgent) EndEpisode(core.Reward)                                         {}

// bestPolicyProvider is implemented by the network population agents (ga, neat):
// a frozen snapshot of the current best member as an observation→action forward.
type bestPolicyProvider interface {
	BestPolicy() func(obs []float64) []float64
}

// chessPolicyProvider is implemented by the search agent (evochess): a frozen
// snapshot of the best member as a state→action move policy (it needs the board,
// not just the observation).
type chessPolicyProvider interface {
	FrozenPolicy() func(core.State) core.Action
}

// bestChessPolicy extracts a frozen best-member policy from a population agent,
// whether it plays by network output (ga/neat) or by search (evochess).
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

// newChessAgent builds the chosen agent sized for chess. evochess is the
// search-based agent (it ignores the observation/action specs and needs a search
// depth); ga and neat reuse the generic network-agent factory.
func newChessAgent(agent string, obs, act core.Spec, population, depth int, seed int64) (core.PopulationAgent, error) {
	if agent == "evochess" {
		c := evochess.DefaultConfig()
		c.Population, c.Depth, c.Seed = population, depth, seed
		return evochess.New(c), nil
	}
	return newPopulationAgent(agent, obs, act, population, seed)
}

func init() {
	var (
		agent       string
		opponent    string
		generations int
		population  int
		maxPlies    int
		depth       int
		seed        int64
		headless    bool
	)
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
			if agent != "ga" && agent != "neat" && agent != "evochess" {
				return fmt.Errorf("chess supports only the ga, neat, and evochess agents")
			}
			if opponent != "random" && opponent != "coevolution" {
				return fmt.Errorf("unknown opponent %q (want random or coevolution)", opponent)
			}
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))
			seed := resolveSeed(cmd, 0, log)

			pop, best, err := evolveChess(agent, opponent == "coevolution", generations, population, maxPlies, depth, seed, log)
			if err != nil {
				return err
			}
			if headless {
				fmt.Printf("best fitness %.3f\n", best)
				return nil
			}
			return watchChess(pop, maxPlies, seed, agent, log)
		},
	}
	cmd.Flags().StringVar(&agent, "agent", "ga", "agent: ga, neat, or evochess")
	cmd.Flags().StringVar(&opponent, "opponent", "coevolution", "opponent: coevolution or random")
	cmd.Flags().IntVar(&generations, "generations", 20, "generations to evolve")
	cmd.Flags().IntVar(&population, "population", 40, "population size")
	cmd.Flags().IntVar(&maxPlies, "max-plies", 60, "maximum plies before a game is drawn")
	cmd.Flags().IntVar(&depth, "depth", 2, "alpha-beta search depth (evochess only)")
	cmd.Flags().Int64Var(&seed, "seed", 0, "run seed (default: random, logged)")
	cmd.Flags().BoolVar(&headless, "headless", false, "train without a window")
	rootCmd.AddCommand(cmd)
}

// evolveChess runs the evolutionary loop and returns the population (with its
// final generation evaluated) and the best fitness reached. When coevolve is
// set, each generation is scored against a frozen champion — the previous
// generation's best — so the opponent strengthens alongside the population.
func evolveChess(agent string, coevolve bool, generations, population, maxPlies, depth int, seed int64, log *slog.Logger) (core.PopulationAgent, float64, error) {
	sample := chess.New(chess.Config{MaxPlies: maxPlies})
	pop, err := newChessAgent(agent, sample.ObservationSpec(), sample.ActionSpec(), population, depth, seed)
	if err != nil {
		return nil, 0, err
	}
	tel := sim.NewTelemetry(generations, log)
	log.Info("evolving", "agent", agent, "opponent", boolPick(coevolve, "coevolution", "random"),
		"generations", generations, "population", population, "seed", seed)

	var champion chess.Policy
	evaluate := func() []float64 {
		opp := champion // snapshot for this generation's factory
		factory := sim.EnvFactory(func() core.MultiEnvironment {
			return sim.AsMulti(func() core.Environment { return chess.New(chess.Config{MaxPlies: maxPlies, Opponent: opp}) })
		})
		return sim.EvaluateParallel(factory, individuals(pop), maxPlies, seed)
	}

	var fitness []float64
	for range generations {
		fitness = evaluate()
		tel.Publish(sim.StatsFrom(pop.Generation(), fitness))
		if coevolve {
			champion = bestChessPolicy(pop)
		}
		pop.Evolve()
	}
	// Evaluate the final (just-evolved) generation so its best is meaningful.
	fitness = evaluate()
	return pop, maxOf(fitness), nil
}

// watchChess opens a window showing the evolved best player (White) against a
// random opponent, looping games.
func watchChess(pop core.PopulationAgent, maxPlies int, seed int64, agent string, log *slog.Logger) error {
	best := bestChessPolicy(pop)
	if best == nil {
		return fmt.Errorf("agent %q cannot expose a best policy", agent)
	}
	env := chess.New(chess.Config{MaxPlies: maxPlies}) // random Black opponent
	caps := runCaps{keymap: onlineKeymap, bounds: boundsOf(env), leader: noLeader, sensors: noSensors}
	return launchGUI(onlineGUI("Galapagos — chess ("+agent+")", env, policyAgent{best}, maxPlies, seed, caps), log)
}

// boolPick returns a when cond is true, else b.
func boolPick(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
