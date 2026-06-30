package packages

import (
	"sort"
	"strings"
)

// CheckMissingExampleImports walks the given workspace dir, parses every
// gno.land/* import declared in its packages, and returns the sorted,
// deduplicated list of imports unreachable via the loader's filesystem
// roots. Used when -no-examples is set to surface broken graphs at startup
// instead of letting them blow up at first query.
//
// Stdlib imports are ignored. The check is FS-only and non-mutating: it
// never reaches the rpc fetcher and never writes to the loader's index or
// tracked sets, so it is safe to call before LoadWorkspace.
//
// Returns nil if workspace is empty.
func CheckMissingExampleImports(l *Loader, workspace string) []string {
	if workspace == "" {
		return nil
	}
	pkgIdx := scanRoot(workspace, nil, l.cfg.Logger)
	seen := map[string]struct{}{}
	for mod, dir := range pkgIdx {
		pkg := &Package{ImportPath: mod, Dir: dir, Kind: KindFS}
		imports, err := pkg.Imports()
		if err != nil {
			l.cfg.Logger.Debug("skipping unreadable package", "dir", dir, "err", err)
			continue
		}
		for _, imp := range imports {
			seen[imp] = struct{}{}
		}
	}
	missing := make([]string, 0, len(seen))
	for imp := range seen {
		if !strings.HasPrefix(imp, "gno.land/") {
			continue
		}
		// Workspace-internal imports are always resolvable by the eager
		// load; LookupFS only covers extra roots and examples.
		if _, ok := pkgIdx[imp]; ok {
			continue
		}
		if l.LookupFS(imp) {
			continue
		}
		missing = append(missing, imp)
	}
	sort.Strings(missing)
	return missing
}
