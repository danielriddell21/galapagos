// Package webui embeds the browser (WebAssembly) demo assets and serves them.
// The native binary embeds and serves these files without any graphics
// dependency, so the browser demo is available from every build — including the
// cross-platform, CGO-free release binary.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// assets holds the web directory. In a fresh checkout it contains only
// index.html; a release (or `just wasm`) adds galapagos.wasm and wasm_exec.js.
//
//go:embed web
var assets embed.FS

// sub returns the embedded web directory rooted at its contents.
func sub() fs.FS {
	f, _ := fs.Sub(assets, "web")
	return f
}

// Bundled reports whether the WebAssembly module was built into this binary.
func Bundled() bool {
	_, err := assets.Open("web/galapagos.wasm")
	return err == nil
}

// Handler serves the embedded demo, tagging .wasm responses with the content
// type required by the browser's streaming instantiation.
func Handler() http.Handler {
	files := http.FileServer(http.FS(sub()))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}
		files.ServeHTTP(w, r)
	})
}
