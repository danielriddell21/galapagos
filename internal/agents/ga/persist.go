package ga

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/danielriddell21/galapagos/internal/core"
)

const schemaVersion = 1

type SavedGenome struct {
	SchemaVersion int       `json:"schema_version"`
	Inputs        int       `json:"inputs"`
	HiddenSize    int       `json:"hidden_size"`
	Outputs       int       `json:"outputs"`
	Generation    int       `json:"generation"`
	Genome        []float64 `json:"genome"`
}

func (p *Population) SaveBest(path string) error {
	sg := SavedGenome{
		SchemaVersion: schemaVersion,
		Inputs:        p.cfg.Inputs,
		HiddenSize:    p.cfg.HiddenSize,
		Outputs:       p.cfg.Outputs,
		Generation:    p.gen,
		Genome:        p.Best(),
	}
	data, err := json.MarshalIndent(sg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal genome: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write genome %q: %w", path, err)
	}
	return nil
}

func LoadGenome(path string) (SavedGenome, error) {
	var sg SavedGenome
	data, err := os.ReadFile(path)
	if err != nil {
		return sg, fmt.Errorf("read genome %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &sg); err != nil {
		return sg, fmt.Errorf("unmarshal genome %q: %w", path, err)
	}
	if sg.SchemaVersion != schemaVersion {
		return sg, fmt.Errorf("genome schema version %d unsupported, want %d", sg.SchemaVersion, schemaVersion)
	}
	if got, want := len(sg.Genome), GenomeLen(sg.Inputs, sg.HiddenSize, sg.Outputs); got != want {
		return sg, fmt.Errorf("genome length %d does not match shape (want %d)", got, want)
	}
	return sg, nil
}

type Driver struct{ ind *individual }

func NewDriver(sg SavedGenome) *Driver {
	cfg := Config{Inputs: sg.Inputs, HiddenSize: sg.HiddenSize, Outputs: sg.Outputs}
	return &Driver{ind: newIndividual(cfg, sg.Genome)}
}

func (d *Driver) Act(s core.State) core.Action { return d.ind.Act(s) }

func (d *Driver) Fitness() core.Reward { return d.ind.Fitness() }

func (d *Driver) SetFitness(r core.Reward) { d.ind.SetFitness(r) }

func (d *Driver) Genome() []float64 { return d.ind.Genome() }

var _ core.Individual = (*Driver)(nil)
