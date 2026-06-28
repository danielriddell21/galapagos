# Contributing to galapagos

## Requirements

* [Go](https://go.dev) (stable — version from `go.mod`)
* [just](https://github.com/casey/just)
* [golangci-lint](https://golangci-lint.run/welcome/install/) (for linting)
* [gremlins](https://github.com/go-gremlins/gremlins) (for mutation testing)

## Development workflow

```
just build      # build the binary
just test       # run unit tests
just gui        # run the Ebiten GUI
just wasm       # build the WebAssembly target
just serve      # serve the wasm build locally
just demos      # regenerate demo assets
```

Run `just --list` to see every recipe. Lint with `golangci-lint run --config .golangci.yml` and run the tests before each commit. CI runs lint + test + build on every push to `trunk` and every pull request targeting `trunk`.

## Project layout

```
cmd/galapagos/   entry point
internal/        implementation packages
configs/         configuration files
docs/            documentation
```

## Commit style

```
type(scope): short imperative description
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`. No period at the end of the subject line; keep it under 72 characters.

## Releases

Releases are triggered by pushing a semver tag — maintainers only. A GitHub Actions workflow runs GoReleaser to build the binaries and update the Homebrew tap; it requires the tap app credentials configured as repository secrets.
