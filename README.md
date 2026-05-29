# Galapagos

A pluggable Go framework for visualizing learning algorithms in real time. The
flagship demo evolves a population of cars that learn to drive a procedurally
generated race track.

Environments, agents, and the renderer are decoupled behind small interfaces, so
any agent can run in any environment. Runs are deterministic: the same seed and
config always produce the same result.

## Build

Requires Go 1.26 (auto-downloaded via the toolchain directive).

```sh
go build ./cmd/galapagos
```

## Usage

Train headlessly and save the best driver:

```sh
galapagos race --headless --config configs/racing.yaml --out best.json
```

Replay a saved driver:

```sh
galapagos replay --genome best.json --seed 42
```

Run the windowed demo (requires a display and OpenGL; build with the `ebiten`
tag):

```sh
go run -tags ebiten ./cmd/galapagos race --config configs/racing.yaml
```

## Layout

- `internal/core` — the interfaces connecting environments, agents, and the renderer
- `internal/sim` — the simulation loop, parallel evaluation, and telemetry
- `internal/envs` — environments (racing; cart-pole and maze)
- `internal/agents` — learning algorithms (genetic algorithm; Q-learning; NEAT)
- `internal/render` — the headless renderer and the Ebiten window
- `cmd/galapagos` — the command-line entrypoint
