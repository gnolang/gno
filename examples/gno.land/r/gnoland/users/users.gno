package users

import (
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/releases"
)

var (
	changelog   = releases.NewChangelog("r/gnoland/users")
	reValidPath = regexp.MustCompile(`^gno\.land/r/gnoland/users/v[^/]+$`)
)

func init() {
	changelog.NewRelease("v1", "/r/gnoland/users/v1", "[Original PR](https://github.com/gnolang/gno/pull/3166)")
}

// AddRelease is intended to be called via a `r/gnoland/user/v*`'s init().
func AddRelease(note string) {
	callerPkgPath := std.PrevRealm().PkgPath()
	if !reValidPath.MatchString(callerPkgPath) {
		return
	}

	ver := strings.TrimPrefix(callerPkgPath, "gno.land/r/gnoland/users/")
	changelog.NewRelease(ver, strings.TrimPrefix(callerPkgPath, "gno.land"), note)
}

func Render(_ string) string {
	return changelog.RenderAsTable(10)
}
