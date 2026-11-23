package gnoweb

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

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

// storeDenyEntries returns the -store-deny entries in the form listings are
// matched in ("acme", "acme/swap"), from whatever an operator types or
// pastes: that form, a gnoweb path ("/r/acme/swap"), a package path
// ("gno.land/r/acme/swap") or a page's address. A deny is an emergency
// switch, so unlike a trusted path, where a typo trusts less, an entry that
// does not name exactly one namespace or package path is an error: it must
// neither fail quietly nor deny more than it names.
func storeDenyEntries(entries []string, domain string) ([]string, error) {
	var out []string
	for _, entry := range entries {
		e := strings.TrimSpace(entry)
		if e == "" {
			continue // the blank after a trailing comma
		}
		pkg, ok := storeDenyEntry(e, domain)
		if !ok {
			return nil, fmt.Errorf("invalid `StoreDeny` entry %q: want a namespace or a package path, such as acme or acme/swap", entry)
		}
		out = append(out, pkg)
	}
	return out, nil
}

// storeDenyEntry returns the namespace or package path e names, if it
// names one.
func storeDenyEntry(e, domain string) (string, bool) {
	// The domain is cut as a whole name: "gno.landscam" does not name scam.
	if rest, ok := strings.CutPrefix(e, domain); ok && (rest == "" || rest[0] == '/') {
		e = rest
	}
	e, _, anchored := strings.Cut(e, "#") // a page's address may end with an anchor
	if strings.Contains(typedPath(e), "//") {
		// An empty name: read as a host first, or dropped with what
		// follows it ("acme//swap.gno" would deny acme).
		return "", false
	}
	_, schemed := cutScheme(e)
	if !schemed && !strings.HasPrefix(e, "/") {
		e = "/" + e
	}
	u, err := weburl.Parse(e)
	if err != nil {
		return "", false
	}
	// A path typed bare is under /r/; a page's address names its tree.
	if tree := u.Path + "/"; !strings.HasPrefix(tree, "/r/") && !strings.HasPrefix(tree, "/p/") && !strings.HasPrefix(tree, "/u/") {
		if schemed {
			return "", false
		}
		u.Path = "/r" + u.Path
	}
	pkg, ok := packagePath(u)
	switch {
	case !ok, pkg == "r", pkg == "p", pkg == "u", !gno.ReGnoUserPkgPath.Matches("gno.land/r/" + pkg):
		return "", false // no package can live there
	case !strings.Contains(pkg, "/"):
		// Anything past a namespace ("scam:app", "scam/App") is a last name
		// weburl dropped: it would deny the whole namespace.
		if anchored || u.File != "" || u.Args != "" || len(u.WebQuery) > 0 || len(u.Query) > 0 {
			return "", false
		}
	case u.File != "" && !isFileName(u.File): // "acme/swap/V2" names no file
		return "", false
	}
	return pkg, true
}

// typedPath is the path of an address as typed, unescaped: past its scheme
// and host, before its arguments and query.
func typedPath(e string) string {
	if rest, ok := cutScheme(e); ok {
		i := strings.IndexByte(rest, '/')
		if i < 0 {
			return ""
		}
		e = rest[i:]
	}
	if i := strings.IndexAny(e, ":$?"); i >= 0 {
		e = e[:i]
	}
	if unescaped, err := url.PathUnescape(e); err == nil {
		return unescaped
	}
	return e
}

// cutScheme returns e past its leading "scheme://", if it starts with one.
func cutScheme(e string) (string, bool) {
	i := strings.Index(e, "://")
	if i <= 0 || !(e[0] >= 'a' && e[0] <= 'z' || e[0] >= 'A' && e[0] <= 'Z') {
		return e, false
	}
	for _, c := range e[1:i] {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.') {
			return e, false
		}
	}
	return e[i+3:], true
}

// isFileName reports a name a package's file can have ("render.gno",
// "README.md"): a base and an extension, of letters, digits, "_", "-" and
// ".", starting with a letter, a digit or "_".
func isFileName(name string) bool {
	ext := path.Ext(name)
	if len(ext) < 2 || len(name) == len(ext) || name[0] == '.' || name[0] == '-' {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

// storeDenied matches a package path, or a namespace, against the entries
// storeDenyEntries returned: an entry denies its own path and all under it.
func storeDenied(entries []string) func(pkg string) bool {
	return newTrustedPaths(entries).contains
}
