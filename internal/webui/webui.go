package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed web
var assets embed.FS

func sub() fs.FS {
	f, _ := fs.Sub(assets, "web")
	return f
}

func Bundled() bool {
	_, err := assets.Open("web/galapagos.wasm")
	return err == nil
}

func Handler() http.Handler {
	files := http.FileServer(http.FS(sub()))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}
		files.ServeHTTP(w, r)
	})
}
