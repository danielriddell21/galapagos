package efficientcube

import (
	"fmt"

	"github.com/danielriddell21/galapagos/internal/nn"
)

func (p *Policy) Save(path string) error {
	if err := p.net.Save(path); err != nil {
		return fmt.Errorf("save policy: %w", err)
	}
	return nil
}

func LoadPolicy(path string) (*Policy, error) {
	m, err := nn.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load policy: %w", err)
	}
	return &Policy{net: m}, nil
}
