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

# Browser build: compiles the windowed demo to WebAssembly and stages web assets.
wasm:
    GOOS=js GOARCH=wasm go build -tags ebiten -o web/galapagos.wasm ./cmd/galapagos
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js
    @echo "serve the web/ directory and open index.html"
