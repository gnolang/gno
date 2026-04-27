package packages

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// CheckMissingExampleImports walks the given workspace dir, parses every
// gno.land/* import declared in its packages, and returns the sorted,
// deduplicated list of imports the loader can't resolve. Used when
// -no-examples is set to surface broken graphs at startup instead of
// letting them blow up at first query.
//
// Stdlib imports are ignored. Imports that successfully resolve via the
// loader (workspace, extra-roots, modcache, RPC) are also ignored.
//
// Returns nil if workspace is empty.
func CheckMissingExampleImports(l *Loader, workspace string) []string {
	if workspace == "" {
		return nil
	}
	pkgIdx := scanRoot(workspace, l.cfg.Logger)
	seen := map[string]struct{}{}
	for _, dir := range pkgIdx {
		for _, imp := range importsInDir(dir) {
			seen[imp] = struct{}{}
		}
	}
	var missing []string
	for imp := range seen {
		if !strings.HasPrefix(imp, "gno.land/") {
			continue
		}
		if gnolang.IsStdlib(imp) {
			continue
		}
		if _, err := l.Resolve(imp); err == nil {
			continue
		}
		missing = append(missing, imp)
	}
	sort.Strings(missing)
	return missing
}

// importsInDir returns the set of import paths declared in dir's non-test
// .gno files. Parses imports only via go/parser.ImportsOnly — gno files are
// Go-syntax-compatible at the import-decl level, so this is safe and avoids
// adding API surface to gnolang.
func importsInDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := map[string]struct{}{}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".gno") {
			continue
		}
		if strings.HasSuffix(name, "_test.gno") || strings.HasSuffix(name, "_filetest.gno") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		af, err := parser.ParseFile(fset, name, body, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, imp := range af.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			out[path] = struct{}{}
		}
	}
	res := make([]string, 0, len(out))
	for k := range out {
		res = append(res, k)
	}
	return res
}
