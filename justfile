# list available recipes
default:
    @just --list

# headless binary (no display dependencies)
[group('build')]
build:
    go build ./cmd/galapagos

# run the full test suite with the race detector
[group('test')]
test:
    go test ./... -race

# golangci-lint (default, CGO-free build)
[group('dev')]
lint:
    golangci-lint run

# format the code
[group('dev')]
fmt:
    golangci-lint fmt

# tidy module dependencies
[group('dev')]
tidy:
    go mod tidy

# full gate: lint + test + build. all must pass before committing
[group('dev')]
ci: lint test build

# windowed demo; requires a display and OpenGL
[group('run')]
gui:
    go run ./cmd/galapagos race --config configs/racing.yaml

# browser demo: compile the windowed demo to WebAssembly into the webui package,
# so it is embedded into the binary and served by `galapagos serve`
[group('build')]
wasm:
    GOOS=js GOARCH=wasm go build -tags ebiten -o internal/webui/web/galapagos.wasm ./cmd/galapagos
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" internal/webui/web/wasm_exec.js

# build the browser demo, then serve it
[group('run')]
serve: wasm
    go run ./cmd/galapagos serve

# regenerate the demo media under docs/demos
[group('run')]
demos:
    # Rendered headlessly through the software renderer: no window, no
    # display, no ebiten build tag. The clips are defined in tools/demogen.
    go run ./tools/demogen
