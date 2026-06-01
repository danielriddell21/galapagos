package main

import (
	"fmt"
	"log/slog"

	"github.com/danielriddell21/galapagos/internal/agents/qlearning"
	"github.com/danielriddell21/galapagos/internal/core"
	"github.com/danielriddell21/galapagos/internal/sim"
)

// envParams describes a run of a single-agent environment (cart-pole or maze)
// with any applicable agent.
type envParams struct {
	title     string
	newSingle func() core.Environment // builds a fresh environment instance
	agent     string                  // ga, neat, or qlearning
	seed      int64
	headless  bool

	generations int // population agents
	population  int
	episodes    int // online agents
	maxSteps    int
	qbins       int // q-learning bins per observation dimension
}

// runEnv dispatches a single-agent environment to the chosen agent, headless or
// windowed. Population agents (ga, neat) train via the multi-adapter as a swarm
// on one shared task; the online agent (qlearning) learns one episode at a time.
func runEnv(p envParams, log *slog.Logger) error {
	sample := p.newSingle()
	obs, act := sample.ObservationSpec(), sample.ActionSpec()

	if p.agent == "qlearning" {
		actions := discreteCount(act)
		if actions == 0 {
			return fmt.Errorf("qlearning needs a discrete-action environment")
		}
		agent := qlearning.New(qlearning.Config{
			Bins: p.qbins, Actions: actions, Alpha: 0.3, Gamma: 0.99,
			Epsilon: 1.0, EpsilonDecay: 0.995, EpsilonMin: 0.01, Seed: p.seed,
		})
		if p.headless {
			rewards := sim.TrainAgent(p.newSingle(), agent, p.episodes, p.maxSteps, p.seed)
			log.Info("trained", "episodes", p.episodes, "states_seen", agent.States(), "final_epsilon", agent.Epsilon())
			fmt.Printf("final episode return %.3f\n", float64(rewards[len(rewards)-1]))
			return nil
		}
		env := p.newSingle()
		caps := runCaps{keymap: onlineKeymap, bounds: boundsOf(env), leader: noLeader, sensors: noSensors}
		return launchGUI(onlineGUI(p.title, env, agent, p.maxSteps, p.seed, caps), log)
	}

	pop, err := newPopulationAgent(p.agent, obs, act, p.population, p.seed)
	if err != nil {
		return err
	}
	factory := func() core.MultiEnvironment { return sim.AsMulti(p.newSingle) }
	if p.headless {
		tel := sim.NewTelemetry(p.generations, log)
		log.Info("training", "agent", p.agent, "generations", p.generations, "population", p.population, "seed", p.seed)
		sim.TrainHeadless(factory, pop, sim.RunConfig{Seed: p.seed, MaxSteps: p.maxSteps, Generations: p.generations}, tel)
		best := sim.EvaluateParallel(factory, individuals(pop), p.maxSteps, p.seed)
		fmt.Printf("best fitness %.3f\n", maxOf(best))
		return nil
	}
	env := sim.AsMulti(p.newSingle)
	caps := runCaps{keymap: swarmKeymap, bounds: boundsOf(env), leader: noLeader, sensors: noSensors}
	return launchGUI(populationGUI(p.title, env, pop, p.maxSteps, p.seed, caps), log)
}

// individuals collects a population's members for evaluation.
func individuals(pop core.PopulationAgent) []core.Individual {
	var out []core.Individual
	for m := range pop.All() {
		out = append(out, m)
	}
	return out
}

// maxOf returns the largest value in xs.
func maxOf(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs {
		m = max(m, x)
	}
	return m
}
