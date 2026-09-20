# galapagos

> *galapagos* — evolution, observed.

[![CI](https://github.com/danielriddell21/galapagos/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/galapagos/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/galapagos/graph/badge.svg)](https://codecov.io/gh/danielriddell21/galapagos)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_galapagos&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_galapagos)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A pluggable Go framework for visualizing learning algorithms in real time. The flagship demo evolves a population of cars that learn to drive a procedurally generated race track.

Environments, agents and the renderer are decoupled behind small interfaces, so any agent can run in any environment. Runs are deterministic from their seed.

## Environments and agents

| Environment | `ga` | `neat` | `qlearning` | `efficientcube` | `evochess` |
|---|---|---|---|---|---|
| `race` | ✅ | ✅ | — | — | — |
| `cartpole` | ✅ | ✅ | ✅ | — | — |
| `maze` | ✅ | ✅ | ✅ | — | — |
| `cube` | — | — | — | ✅ | — |
| `flappy` | ✅ | ✅ | — | — | — |
| `chess` | ✅ | ✅ | — | — | ✅ |

## Install

On macOS, the Homebrew **cask** installs the full binary — the native window plus every CLI command:

```bash
brew install --cask danielriddell21/tap/galapagos
```

On Linux (or macOS without the native window), the Homebrew **formula** installs the cross-platform CLI:

```bash
brew install danielriddell21/tap/galapagos
```

### From source

Requires Go 1.26 (auto-downloaded via the toolchain directive).

```sh
go build ./cmd/galapagos
```

### The browser demo

`galapagos serve` compiles the windowed demo to WebAssembly and serves it, so
it runs anywhere a browser does — no native OpenGL involved.

It is built when you ask for it rather than shipped inside the binary: the
WebAssembly build is 24MB, and carrying it in every archive for every platform
made the downloads four times larger for the sake of one command. So `serve`
needs Go and a checkout of this repository:

```sh
git clone https://github.com/danielriddell21/galapagos
cd galapagos
just serve
```

<details>
<summary>Linux: OpenGL/X11 libraries</summary>

Building the native window from source on Linux (`go build -tags ebiten ./cmd/galapagos`)
needs OpenGL/X11. On Debian/Ubuntu:

```sh
sudo apt install libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev
```
</details>

## Quick start

```sh
# Train the racing demo headlessly and save the best driver
galapagos race --headless --config configs/racing.yaml --out best.json

# Replay a saved driver — reproduces the trained result exactly
galapagos replay --genome best.json --seed 42

# Any applicable agent runs on any environment
galapagos cartpole --headless --agent qlearning
```

## Documentation

Full documentation lives in the [galapagos wiki](https://github.com/danielriddell21/galapagos/wiki) — every environment and agent, the windowed and browser demos, and the architecture.
