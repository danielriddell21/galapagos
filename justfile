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

# compile the windowed demo to WebAssembly and serve it. The build happens
# inside the command, into a temporary directory: the demo is not shipped in
# the binary, so there is nothing to stage here first.
[group('run')]
serve:
    go run ./cmd/galapagos serve

# regenerate the demo media under docs/demos
[group('run')]
demos:
    # Rendered headlessly through the software renderer: no window, no
    # display, no ebiten build tag. The clips are defined in tools/demogen.
    go run ./tools/demogen
