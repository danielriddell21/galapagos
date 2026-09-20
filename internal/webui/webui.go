// Package webui serves the browser build of the windowed demo.
//
// The WebAssembly binary is not embedded. It is 24MB, and embedding it put
// that in every archive for every platform for the sake of one command — the
// CLI archives were four times the size of the comparable games because of it.
// It is compiled when somebody asks for it instead, which also means the demo
// can never be a stale copy of the source sitting beside it.
//
// The page that loads it is embedded, because it is a few hundred bytes and
// nothing generates it.
package webui

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// modulePath is the module the demo is built from, used to tell "you are not
// in the checkout" apart from every other reason a build can fail.
const modulePath = "github.com/danielriddell21/galapagos"

//go:embed web/index.html
var page embed.FS

// Build compiles the browser demo into dir, next to the page that loads it.
//
// Go is required, and so is the module: the demo is the same program compiled
// for a different target, so building it means building this repository. A
// binary installed on its own cannot do it, which is why the error says so
// rather than reporting a missing file.
func Build(ctx context.Context, dir string) error {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("the browser demo is compiled on demand, and needs Go on PATH: %w", err)
	}
	if err := inModule(ctx, goBin); err != nil {
		return err
	}

	build := exec.CommandContext(ctx, goBin, "build", "-tags", "ebiten",
		"-o", filepath.Join(dir, "galapagos.wasm"), "./cmd/galapagos")
	build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("compiling the browser demo: %w\n%s", err, out)
	}

	// wasm_exec.js is the glue the Go toolchain ships, and it has to match the
	// toolchain that produced the .wasm beside it.
	root, err := goEnv(ctx, goBin, "GOROOT")
	if err != nil {
		return err
	}
	glue, err := os.ReadFile(filepath.Join(root, "lib", "wasm", "wasm_exec.js"))
	if err != nil {
		return fmt.Errorf("reading wasm_exec.js from the Go toolchain: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "wasm_exec.js"), glue, 0o600); err != nil {
		return fmt.Errorf("writing wasm_exec.js: %w", err)
	}

	index, err := page.ReadFile("web/index.html")
	if err != nil {
		return fmt.Errorf("reading the demo page: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), index, 0o600); err != nil {
		return fmt.Errorf("writing the demo page: %w", err)
	}
	return nil
}

// inModule reports whether the working directory is inside this module.
func inModule(ctx context.Context, goBin string) error {
	found, err := goEnv(ctx, goBin, "GOMOD")
	if err != nil || found == "" || found == os.DevNull {
		return fmt.Errorf(
			"the browser demo is compiled on demand, so it has to be run from a %s checkout;\n"+
				"  clone the repository and run `galapagos serve` there", modulePath)
	}

	list := exec.CommandContext(ctx, goBin, "list", "-m")
	out, err := list.Output()
	if err != nil || strings.TrimSpace(string(out)) != modulePath {
		return fmt.Errorf(
			"the browser demo is compiled on demand, so it has to be run from a %s checkout;\n"+
				"  this is %s", modulePath, strings.TrimSpace(string(out)))
	}
	return nil
}

func goEnv(ctx context.Context, goBin, name string) (string, error) {
	out, err := exec.CommandContext(ctx, goBin, "env", name).Output()
	if err != nil {
		return "", fmt.Errorf("reading %s from the Go toolchain: %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Handler serves a directory Build has written.
func Handler(dir string) http.Handler {
	return handlerFor(os.DirFS(dir))
}

func handlerFor(files fs.FS) http.Handler {
	server := http.FileServer(http.FS(files))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}
		server.ServeHTTP(w, r)
	})
}
