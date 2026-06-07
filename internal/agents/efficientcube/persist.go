package efficientcube

import "github.com/danielriddell21/galapagos/internal/nn"

// Save writes the trained policy network to path as JSON.
func (p *Policy) Save(path string) error { return p.net.Save(path) }

// LoadPolicy reads a policy network from path.
func LoadPolicy(path string) (*Policy, error) {
	m, err := nn.Load(path)
	if err != nil {
		return nil, err
	}
	return &Policy{net: m}, nil
}
