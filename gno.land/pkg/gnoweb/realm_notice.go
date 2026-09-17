package gnoweb

import (
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

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
