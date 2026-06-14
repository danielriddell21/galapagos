# Galapagos developer tasks.

# Headless binary (no display dependencies).
build:
    go build ./cmd/galapagos

# Run the full test suite with the race detector.
test:
    go test ./... -race

# Windowed demo; requires a display and OpenGL.
gui:
    go run -tags ebiten ./cmd/galapagos race --config configs/racing.yaml

# Browser demo: compile the windowed demo to WebAssembly into the webui package,
# so it is embedded into the binary and served by `galapagos serve`.
wasm:
    GOOS=js GOARCH=wasm go build -tags ebiten -o internal/webui/web/galapagos.wasm ./cmd/galapagos
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" internal/webui/web/wasm_exec.js

# Build the browser demo, then serve it.
serve: wasm
    go run ./cmd/galapagos serve

# Record a demo GIF per GUI tool (requires a display + OpenGL). Each tool is run
# explicitly and writes one GIF into docs/demos/.
demos:
    mkdir -p docs/demos
    go run -tags ebiten ./cmd/galapagos race --record docs/demos/race.gif --record-frames 220 --seed 7
    go run -tags ebiten ./cmd/galapagos cartpole --agent ga --record docs/demos/cartpole.gif --record-frames 180 --seed 7
    go run -tags ebiten ./cmd/galapagos maze --agent qlearning --record docs/demos/maze.gif --record-frames 180 --seed 7
    go run -tags ebiten ./cmd/galapagos cube --iters 1500 --record docs/demos/cube.gif --record-frames 200 --seed 7
    go run -tags ebiten ./cmd/galapagos flappy --agent ga --record docs/demos/flappy.gif --record-frames 130 --seed 7
    go run -tags ebiten ./cmd/galapagos chess --agent ga --opponent coevolution --record docs/demos/chess.gif --record-frames 95 --seed 7
