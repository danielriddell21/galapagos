package webui

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// A browser fetching the module gets it rejected outright unless the server
// says it is WebAssembly, so the one header this handler adds is the one that
// makes the demo work at all.
func TestHandlerLabelsTheModuleAsWebAssembly(t *testing.T) {
	files := fstest.MapFS{
		"index.html":     {Data: []byte("<html></html>")},
		"galapagos.wasm": {Data: []byte("\x00asm")},
		"wasm_exec.js":   {Data: []byte("// glue")},
	}
	server := httptest.NewServer(handlerFor(files))
	t.Cleanup(server.Close)

	tests := []struct {
		path string
		want string
	}{
		{"/galapagos.wasm", "application/wasm"},
		{"/wasm_exec.js", "text/javascript"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := http.Get(server.URL + tt.path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = resp.Body.Close() }()
			_, _ = io.Copy(io.Discard, resp.Body)

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d", resp.StatusCode)
			}
			if got := resp.Header.Get("Content-Type"); !strings.Contains(got, tt.want) {
				t.Errorf("Content-Type = %q, want it to mention %q", got, tt.want)
			}
		})
	}
}

// The demo is the same program compiled for a different target, so building it
// means building this module. Run from anywhere else it has to say so — the
// alternative is a confusing failure from the Go toolchain about a package
// pattern that matches nothing.
func TestBuildRefusesOutsideTheModule(t *testing.T) {
	t.Chdir(t.TempDir())

	err := Build(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("Build() outside the module should fail")
	}
	for _, want := range []string{"compiled on demand", modulePath, "checkout"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to mention %q", err, want)
		}
	}
}

// The page is the one part that is still embedded, because nothing generates
// it. Build writes it out beside what it compiles.
func TestPageLoadsTheModuleBesideIt(t *testing.T) {
	index, err := page.ReadFile("web/index.html")
	if err != nil {
		t.Fatalf("the demo page is not embedded: %v", err)
	}
	for _, want := range []string{"galapagos.wasm", "wasm_exec.js"} {
		if !strings.Contains(string(index), want) {
			t.Errorf("the page does not reference %q", want)
		}
	}
}

// Whatever Build writes, Handler has to be able to serve.
func TestHandlerServesWhatBuildWrites(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"index.html", "galapagos.wasm", "wasm_exec.js"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	server := httptest.NewServer(Handler(dir))
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("serving the demo directory: status = %d", resp.StatusCode)
	}
}
