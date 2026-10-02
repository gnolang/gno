package gnoweb

import "strings"

// trustedPaths is keyed without the "/r/", "/p/" or "/u/" prefix so one
// entry covers every tree. An entry trusts its own path and everything under
// it, and "*" trusts every path, for a chain whose packages are all the
// operator's own, such as gnodev's.
type trustedPaths map[string]struct{}

func newTrustedPaths(entries []string) trustedPaths {
	set := make(trustedPaths, len(entries))
	for _, e := range entries {
		if e = strings.Trim(e, " /"); e != "" {
			set[e] = struct{}{}
		}
	}
	return set
}

// contains reports whether pkg or one of its parent paths is trusted.
func (t trustedPaths) contains(pkg string) bool {
	if _, all := t["*"]; all {
		return true
	}
	pkg = strings.TrimSuffix(pkg, "/")
	for {
		if _, ok := t[pkg]; ok {
			return true
		}
		i := strings.LastIndexByte(pkg, '/')
		if i < 0 {
			return false
		}
		pkg = pkg[:i]
	}
}

// malformedTrustedPaths returns the entries that can match no package: one
// that keeps a "r/", "p/", "u/" or domain prefix, or has an uppercase letter,
// which package paths never do. They stay in the list, matching nothing, so
// a typo trusts less, never more.
func malformedTrustedPaths(entries []string) []string {
	var bad []string
	for _, e := range entries {
		e = strings.Trim(e, " /")
		switch {
		case strings.HasPrefix(e, "r/"), strings.HasPrefix(e, "p/"), strings.HasPrefix(e, "u/"),
			strings.HasPrefix(e, "gno.land/"), strings.ToLower(e) != e:
			bad = append(bad, e)
		}
	}
	return bad
}
