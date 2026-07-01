package nn

import (
	"encoding/json"
	"fmt"
	"os"
)

const schemaVersion = 1

type Saved struct {
	SchemaVersion int         `json:"schema_version"`
	Sizes         []int       `json:"sizes"`
	Hidden        Activation  `json:"hidden"`
	Output        Activation  `json:"output"`
	Weights       [][]float64 `json:"weights"`
	Biases        [][]float64 `json:"biases"`
}

func (m *MLP) Snapshot() Saved {
	s := Saved{SchemaVersion: schemaVersion, Sizes: m.sizes, Hidden: m.hidden, Output: m.output}
	for _, ly := range m.layers {
		s.Weights = append(s.Weights, append([]float64(nil), ly.w...))
		s.Biases = append(s.Biases, append([]float64(nil), ly.b...))
	}
	return s
}

func FromSnapshot(s Saved) (*MLP, error) {
	if s.SchemaVersion != schemaVersion {
		return nil, fmt.Errorf("nn: schema version %d unsupported, want %d", s.SchemaVersion, schemaVersion)
	}
	m := New(Config{Sizes: s.Sizes, Hidden: s.Hidden, Output: s.Output})
	if len(s.Weights) != len(m.layers) {
		return nil, fmt.Errorf("nn: snapshot has %d layers, want %d", len(s.Weights), len(m.layers))
	}
	for l := range m.layers {
		if len(s.Weights[l]) != len(m.layers[l].w) || len(s.Biases[l]) != len(m.layers[l].b) {
			return nil, fmt.Errorf("nn: layer %d shape mismatch", l)
		}
		copy(m.layers[l].w, s.Weights[l])
		copy(m.layers[l].b, s.Biases[l])
	}
	return m, nil
}

func (m *MLP) Save(path string) error {
	data, err := json.Marshal(m.Snapshot())
	if err != nil {
		return fmt.Errorf("nn: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("nn: write %q: %w", path, err)
	}
	return nil
}

func Load(path string) (*MLP, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("nn: read %q: %w", path, err)
	}
	var s Saved
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("nn: unmarshal %q: %w", path, err)
	}
	return FromSnapshot(s)
}
