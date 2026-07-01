package efficientcube

import (
	"runtime"
	"sync"

	rubix "github.com/danielriddell21/rubix/pkg/cube"
)

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
