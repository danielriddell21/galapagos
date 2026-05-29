package sim

import (
	"time"

	"github.com/danielriddell21/galapagos/internal/core"
)

// RunConfig parameters one training run. A run is fully determined by Seed plus
// these values together with the agent and environment configuration.
type RunConfig struct {
	Seed        int64
	MaxSteps    int
	Generations int
}

// TrainHeadless evolves pop for cfg.Generations using parallel evaluation and
// publishes per-generation statistics to tel (which may be nil). This is the
// turbo/batch path: no rendering, maximum speed.
func TrainHeadless(newEnv EnvFactory, pop core.PopulationAgent, cfg RunConfig, tel *Telemetry) {
	for range cfg.Generations {
		fitness := EvaluateParallel(newEnv, collect(pop), cfg.MaxSteps, cfg.Seed)
		if tel != nil {
			tel.Publish(StatsFrom(pop.Generation(), fitness))
		}
		pop.Evolve()
	}
}

// Pace invokes step exactly steps times, sleeping so that successive calls are
// spaced by interval. It models the real-time render loop; under testing/synctest
// its sleeps advance the fake clock, making frame-paced behavior deterministic.
func Pace(steps int, interval time.Duration, step func()) {
	for i := range steps {
		step()
		if i < steps-1 {
			time.Sleep(interval)
		}
	}
}
