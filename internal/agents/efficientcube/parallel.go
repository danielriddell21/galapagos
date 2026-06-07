package efficientcube

import (
	"runtime"
	"sync"

	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

// SolveBatch solves every cube with beam search concurrently, fanning the work
// across a pool sized to runtime.GOMAXPROCS(0). Each solve is independent and its
// result is written by index, so the output is identical regardless of the worker
// count. Solving only reads the policy weights, so concurrent use is safe.
func (p *Policy) SolveBatch(cubes []rubix.Cube, cfg BeamConfig) []Result {
	results := make([]Result, len(cubes))
	sem := make(chan struct{}, max(runtime.GOMAXPROCS(0), 1))
	var wg sync.WaitGroup

	for i := range cubes {
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			results[i] = p.Solve(cubes[i], cfg)
		})
	}
	wg.Wait()
	return results
}
