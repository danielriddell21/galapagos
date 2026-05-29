// Package config loads and validates run configuration from YAML. A run is
// fully determined by these values plus the seed, satisfying the project's
// reproducibility requirement.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Racing is the configuration for the racing demo and its genetic algorithm.
// Field names match the YAML keys in configs/racing.yaml.
type Racing struct {
	Agent         string  `yaml:"agent"` // "ga" or "neat"
	Population    int     `yaml:"population"`
	EliteFraction float64 `yaml:"elite_fraction"`
	MutationRate  float64 `yaml:"mutation_rate"`
	MutationStd   float64 `yaml:"mutation_std"`
	HiddenSize    int     `yaml:"hidden_size"`
	Rays          int     `yaml:"rays"`
	MaxSteps      int     `yaml:"max_steps"`
	Generations   int     `yaml:"generations"`
	Seed          int64   `yaml:"seed"`
}

// DefaultRacing returns the configuration documented in the spec.
func DefaultRacing() Racing {
	return Racing{
		Agent:         "ga",
		Population:    100,
		EliteFraction: 0.1,
		MutationRate:  0.05,
		MutationStd:   0.2,
		HiddenSize:    8,
		Rays:          7,
		MaxSteps:      2000,
		Generations:   200,
		Seed:          42,
	}
}

// Validate reports the first invalid field, if any.
func (c Racing) Validate() error {
	switch {
	case c.Agent != "ga" && c.Agent != "neat":
		return fmt.Errorf(`agent must be "ga" or "neat", got %q`, c.Agent)
	case c.Population < 2:
		return fmt.Errorf("population must be at least 2, got %d", c.Population)
	case c.EliteFraction < 0 || c.EliteFraction >= 1:
		return fmt.Errorf("elite_fraction must be in [0,1), got %g", c.EliteFraction)
	case c.MutationRate < 0 || c.MutationRate > 1:
		return fmt.Errorf("mutation_rate must be in [0,1], got %g", c.MutationRate)
	case c.HiddenSize < 1:
		return fmt.Errorf("hidden_size must be positive, got %d", c.HiddenSize)
	case c.Rays < 1:
		return fmt.Errorf("rays must be positive, got %d", c.Rays)
	case c.MaxSteps < 1:
		return fmt.Errorf("max_steps must be positive, got %d", c.MaxSteps)
	}
	return nil
}

// LoadRacing reads and validates a racing config from path. Missing fields keep
// their default values.
func LoadRacing(path string) (Racing, error) {
	c := DefaultRacing()
	data, err := os.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("read config %q: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("parse config %q: %w", path, err)
	}
	if err := c.Validate(); err != nil {
		return c, fmt.Errorf("invalid config %q: %w", path, err)
	}
	return c, nil
}
