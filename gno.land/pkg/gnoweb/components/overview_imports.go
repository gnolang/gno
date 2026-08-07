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

// stdlibSourceBase is where the Gno standard library lives. Stdlibs ship with
// the node instead of being deployed on chain, so they have no package page to
// link to and the source has to be reached upstream.
const stdlibSourceBase = "https://github.com/gnolang/gno/tree/master/gnovm/stdlibs/"

func buildImportLink(p, kind, domain string) string {
	switch kind {
	case "package", "realm":
		return strings.TrimPrefix(p, domain)
	case "stdlib":
		return stdlibSourceBase + p
	}
	return ""
}
