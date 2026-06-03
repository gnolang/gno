package components

import (
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

// parseImports extracts deduplicated import paths from .gno sources.
// Uses parser.ImportsOnly which stops after the imports block and tolerates
// .gno-only syntax that would otherwise fail a full parse.
func parseImports(sources map[string][]byte, domain string) []ImportLink {
	if len(sources) == 0 {
		return nil
	}

	seen := map[string]string{} // path -> kind
	fset := token.NewFileSet()
	for _, content := range sources {
		f, err := parser.ParseFile(fset, "", content, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if _, ok := seen[p]; !ok {
				seen[p] = classifyImport(p, domain)
			}
		}
	}
	if len(seen) == 0 {
		return nil
	}

	out := make([]ImportLink, 0, len(seen))
	for p, kind := range seen {
		out = append(out, ImportLink{
			Path: p,
			Kind: kind,
			Link: buildImportLink(p, kind, domain),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func classifyImport(p, domain string) string {
	switch {
	case strings.HasPrefix(p, domain+"/p/"):
		return "package"
	case strings.HasPrefix(p, domain+"/r/"):
		return "realm"
	case strings.Contains(p, "."):
		return "external"
	default:
		return "stdlib"
	}
}

func buildImportLink(p, kind, domain string) string {
	if kind == "package" || kind == "realm" {
		return strings.TrimPrefix(p, domain)
	}
	return ""
}
