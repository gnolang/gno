package components

import "strings"

// buildImports turns the package's import paths — already deduplicated and
// sorted by vm/qdoc — into rendered dependency links.
func buildImports(paths []string, domain string) []ImportLink {
	if len(paths) == 0 {
		return nil
	}
	out := make([]ImportLink, 0, len(paths))
	for _, p := range paths {
		kind := classifyImport(p, domain)
		out = append(out, ImportLink{
			Path: p,
			Kind: kind,
			Link: buildImportLink(p, kind, domain),
		})
	}
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
