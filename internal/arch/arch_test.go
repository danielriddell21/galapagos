package arch

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func importsOf(t *testing.T, root string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		// The rules constrain the production import graph; integration tests may
		// freely combine packages.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		rel, _ := filepath.Rel(root, path)
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			out[rel] = append(out[rel], p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return out
}

func TestDecouplingRules(t *testing.T) {
	const mod = "github.com/danielriddell21/galapagos/internal"
	imports := importsOf(t, "..") // run from internal/arch, so ".." is internal/

	for file, imps := range imports {
		under := func(prefix string) bool { return strings.HasPrefix(file, prefix) }
		for _, imp := range imps {
			switch {
			case imp == "math/rand":
				t.Errorf("%s imports legacy math/rand; use math/rand/v2", file)
			case under("envs"+string(filepath.Separator)) && strings.Contains(imp, "hajimehoshi"):
				t.Errorf("%s (environment) imports Ebiten", file)
			case under("envs"+string(filepath.Separator)) && strings.HasPrefix(imp, mod+"/agents"):
				t.Errorf("%s (environment) imports an agent: %s", file, imp)
			case under("envs"+string(filepath.Separator)) && strings.HasPrefix(imp, mod+"/render"):
				t.Errorf("%s (environment) imports the renderer: %s", file, imp)
			case under("agents"+string(filepath.Separator)) && strings.HasPrefix(imp, mod+"/envs"):
				t.Errorf("%s (agent) imports an environment: %s", file, imp)
			case under("agents"+string(filepath.Separator)) && strings.Contains(imp, "hajimehoshi"):
				t.Errorf("%s (agent) imports Ebiten", file)
			}
		}
	}
}
