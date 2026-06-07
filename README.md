# Galapagos

[![CI](https://github.com/danielriddell21/galapagos/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/galapagos/actions/workflows/ci.yaml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_galapagos&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_galapagos)
[![codecov](https://codecov.io/gh/danielriddell21/galapagos/graph/badge.svg)](https://codecov.io/gh/danielriddell21/galapagos)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A pluggable Go framework for visualizing learning algorithms in real time. The
flagship demo evolves a population of cars that learn to drive a procedurally
generated race track.

Environments, agents, and the renderer are decoupled behind small interfaces, so
any agent can run in any environment. Runs are deterministic from their seed: a
seed is chosen at random when none is given and logged, and passing `--seed N`
(or a config seed) reproduces a run exactly.

| environment | ga | neat | qlearning |
|-------------|----|------|-----------|
| racing      | ✅ | ✅   | —          |
| cartpole    | ✅ | ✅   | ✅         |
| maze        | ✅ | ✅   | ✅         |

The `cube` environment wraps the Rubik's cube model from
[`rubix`](https://github.com/danielriddell21/rubix) and is solved with
**EfficientCube** (Takano, 2023): a neural policy trained by self-supervision —
predicting the move that reverses each scramble step — then solving new scrambles
with beam search. A small pure-Go/CPU net reliably solves shallow-to-moderate
scrambles; deep (≈20-move) scrambles are best-effort. The trainable MLP lives in
`internal/nn`.

## Install

On macOS, the Homebrew **cask** installs the full binary — the native window plus
every CLI command and the browser demo:

```bash
brew install --cask danielriddell21/tap/galapagos
```

On Linux (or macOS without the native window), the Homebrew **formula** installs
the cross-platform CLI; the browser demo rides along via `galapagos serve`:

```bash
brew install danielriddell21/tap/galapagos
```

The native window is macOS-only (it uses Metal, which needs no extra libraries).
On any platform, `galapagos serve` opens the WebAssembly demo in your browser.

## Build

Requires Go 1.26 (auto-downloaded via the toolchain directive).

```sh
go build ./cmd/galapagos
```

## Usage

Train the racing demo headlessly and save the best driver:

```sh
galapagos race --headless --config configs/racing.yaml --out best.json
```

Replay a saved driver (reproduces the trained result exactly):

```sh
galapagos replay --genome best.json --seed 42
```

Any applicable agent runs on any environment via `--agent`:

```sh
galapagos race --headless --config configs/racing-neat.yaml  # NEAT evolves topology
galapagos cartpole --headless --agent ga                     # GA balances a pole
galapagos cartpole --headless --agent qlearning              # tabular Q-learning balances a pole
galapagos maze --headless --agent qlearning                  # Q-learning solves a maze
galapagos cube --headless --eval-depth 6                     # EfficientCube learns to solve a cube
```

### Windowed demo

Every environment has a live, watchable window (requires a display and OpenGL,
built with the `ebiten` tag). Drop `--headless` to open it:

```sh
just gui                                       # racing demo
go run -tags ebiten ./cmd/galapagos cartpole   # cart-pole swarm
go run -tags ebiten ./cmd/galapagos maze --agent qlearning
```

The current controls are listed in the window's top-right overlay: `space`
pause, `f` follow/fit camera, `+`/`-` speed, `r` regenerate (new random seed),
`s` save best, `l` load best, `d` toggle sensor rays.

### Browser demo

The same window compiled to WebAssembly runs in a browser, so no native graphics
libraries are needed. The binary embeds it; `galapagos serve` (or `just serve`)
hosts it and opens a browser.

For a browser build, `just wasm` compiles the demo to WebAssembly and stages the
`web/` directory.

## Layout

- `internal/core` — the interfaces connecting environments, agents, and the renderer
- `internal/sim` — the simulation loop, parallel evaluation, and telemetry
- `internal/envs` — environments (racing; cart-pole and maze)
- `internal/agents` — learning algorithms (genetic algorithm; Q-learning; NEAT)
- `internal/render` — the headless renderer and the Ebiten window
- `cmd/galapagos` — the command-line entrypoint
