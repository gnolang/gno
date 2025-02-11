package users

import (
	"strings"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/releases"

	"gno.land/r/gov/dao/bridge"
)

var changelog = releases.NewChangelog("r/gnoland/users")

const usersPrefix = "gno.land/r/gnoland/users/"

func init() {
	changelog.NewRelease("v1", "/r/gnoland/users/v1", "[Original PR](https://github.com/gnolang/gno/pull/3166)")
}

func Render(_ string) string {
	return changelog.RenderAsTable(10)
}

// NewAddReleaseExecutor allows a GovDAO proposal to add a release to the changelog
func NewAddReleaseExecutor(newVerPkgPath, note string) dao.Executor {
	if !strings.HasPrefix(newVerPkgPath, usersPrefix) {
		panic("invalid version pkgpath")
	}

	cb := func() error {
		ver := strings.TrimPrefix(newVerPkgPath, usersPrefix)
		changelog.NewRelease(ver, strings.TrimPrefix(newVerPkgPath, "gno.land"), note)
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}
