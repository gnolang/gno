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

// trustedPathProblem says why an operator-typed entry cannot match the
// package it names, or "" if it can. Such an entry is kept and trusts no
// intended package, so a typo trusts less, never more.
func trustedPathProblem(entry string) string {
	e := strings.Trim(entry, " /")
	switch {
	case e != "*" && strings.Contains(e, "*"):
		return "contains *: only a bare * is a wildcard, and it trusts every path"
	case strings.Contains(e, "."):
		return "contains a dot: drop the domain"
	case strings.HasPrefix(e, "r/"), strings.HasPrefix(e, "p/"), strings.HasPrefix(e, "u/"):
		return "starts with r/, p/ or u/: drop the prefix"
	case e != strings.ToLower(e):
		return "has uppercase letters: package paths are lowercase"
	}
	return ""
}
