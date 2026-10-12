package gnoweb

import (
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// RealmNoticeHeader is set to "community" on markdown responses of pages that
// show the realm notice, whose body is served verbatim.
const RealmNoticeHeader = "X-Gnoweb-Realm-Notice"

// trustedPaths is keyed without the "/r/" or "/p/" prefix so one entry
// covers both trees. An entry trusts its own path and everything under it.
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

// trustedPathProblem says why an operator-typed entry cannot match the
// package it names, or "" if it can. Such an entry is kept and trusts no
// intended package, so the notice fails closed.
func trustedPathProblem(entry string) string {
	e := strings.Trim(entry, " /")
	switch {
	case strings.Contains(e, "."):
		return "contains a dot: drop the domain"
	case strings.HasPrefix(e, "r/"), strings.HasPrefix(e, "p/"), strings.HasPrefix(e, "u/"):
		return "starts with r/, p/ or u/: drop the prefix"
	case e != strings.ToLower(e):
		return "has uppercase letters: package paths are lowercase"
	}
	return ""
}

// contains reports whether pkg or one of its parent paths is trusted.
func (t trustedPaths) contains(pkg string) bool {
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

// showRealmNotice reports whether u is a package or user page outside the
// trusted paths. A user page renders that user's home realm. The bare "/r/",
// "/p/" and "/u/" listings are none of these.
func (h *HTTPHandler) showRealmNotice(u *weburl.GnoURL) bool {
	if !h.Static.RealmNotice.Enabled() || !(u.IsRealm() || u.IsPure() || u.IsUser()) {
		return false
	}
	pkg := u.Path[3:] // skip "/r/", "/p/" or "/u/"
	return pkg != "" && !h.trusted.contains(pkg)
}
