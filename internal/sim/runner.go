package sim

import (
	"time"

	"github.com/danielriddell21/galapagos/internal/core"
)

type RunConfig struct {
	Seed        int64
	MaxSteps    int
	Generations int
}

func TrainHeadless(newEnv EnvFactory, pop core.PopulationAgent, cfg RunConfig, tel *Telemetry) {
	for range cfg.Generations {
		fitness := EvaluateParallel(newEnv, collect(pop), cfg.MaxSteps, cfg.Seed)
		if tel != nil {
			tel.Publish(StatsFrom(pop.Generation(), fitness))
		}
		pop.Evolve()
	}
}

func Pace(steps int, interval time.Duration, step func()) {
	for i := range steps {
		step()
		if i < steps-1 {
			time.Sleep(interval)
		}
	}
}
