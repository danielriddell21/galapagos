package cli

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/danielriddell21/galapagos/internal/agents/qlearning"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/gui"
	"github.com/danielriddell21/galapagos/internal/sim"
)

type envParams struct {
	title     string
	newSingle func() core.Environment
	agent     string
	seed      int64
	headless  bool

	generations int
	population  int
	episodes    int
	maxSteps    int
	qbins       int
	qByState    bool
}

func runEnv(p envParams, log *slog.Logger) error {
	if p.headless {
		return trainEnv(p, log)
	}
	cfg, err := envConfig(p)
	if err != nil {
		return err
	}
	if err := gui.Run(cfg, log); err != nil {
		return fmt.Errorf("run gui: %w", err)
	}
	return nil
}

// envConfig wires an environment and its agent into a run the display can
// drive. It is display-free, so the same wiring backs the window and the
// headless recordings in tools/demogen.
func envConfig(p envParams) (gui.Config, error) {
	sample := p.newSingle()
	obs, act := sample.ObservationSpec(), sample.ActionSpec()

	if p.agent == "qlearning" {
		agent, err := newQAgent(p, act)
		if err != nil {
			return gui.Config{}, err
		}
		env := p.newSingle()
		caps := runCaps{keymap: onlineKeymap, bounds: boundsOf(env), leader: noLeader, sensors: noSensors}
		return onlineGUI(p.title, env, agent, p.maxSteps, p.seed, caps), nil
	}

	pop, err := newPopulationAgent(p.agent, obs, act, p.population, p.seed)
	if err != nil {
		return gui.Config{}, err
	}
	env := sim.AsMulti(p.newSingle)
	caps := runCaps{keymap: swarmKeymap, bounds: boundsOf(env), leader: noLeader, sensors: noSensors}
	return populationGUI(p.title, env, pop, p.maxSteps, p.seed, caps), nil
}

// trainEnv runs the environment to completion without a display and reports
// what it learned.
func trainEnv(p envParams, log *slog.Logger) error {
	sample := p.newSingle()
	obs, act := sample.ObservationSpec(), sample.ActionSpec()

	if p.agent == "qlearning" {
		agent, err := newQAgent(p, act)
		if err != nil {
			return err
		}
		rewards := sim.TrainAgent(p.newSingle(), agent, p.episodes, p.maxSteps, p.seed)
		log.Info("trained", "episodes", p.episodes, "states_seen", agent.States(), "final_epsilon", agent.Epsilon())
		fmt.Printf("final episode return %.3f\n", float64(rewards[len(rewards)-1]))
		return nil
	}

	pop, err := newPopulationAgent(p.agent, obs, act, p.population, p.seed)
	if err != nil {
		return err
	}
	factory := func() core.MultiEnvironment { return sim.AsMulti(p.newSingle) }
	tel := sim.NewTelemetry(p.generations, log)
	log.Info("training", "agent", p.agent, "generations", p.generations, "population", p.population, "seed", p.seed)
	sim.TrainHeadless(factory, pop, sim.RunConfig{Seed: p.seed, MaxSteps: p.maxSteps, Generations: p.generations}, tel)
	best := sim.EvaluateParallel(factory, individuals(pop), p.maxSteps, p.seed)
	fmt.Printf("best fitness %.3f\n", slices.Max(best))
	return nil
}

func newQAgent(p envParams, act core.Spec) (*qlearning.Agent, error) {
	actions := discreteCount(act)
	if actions == 0 {
		return nil, fmt.Errorf("qlearning needs a discrete-action environment")
	}
	return qlearning.New(qlearning.Config{
		Bins: p.qbins, Actions: actions, Alpha: 0.3, Gamma: 0.99,
		Epsilon: 1.0, EpsilonDecay: 0.995, EpsilonMin: 0.01, Seed: p.seed,
		KeyByState: p.qByState,
	}), nil
}

func individuals(pop core.PopulationAgent) []core.Individual {
	return slices.Collect(pop.All())
}
