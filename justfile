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
