package users

import (
	"chain/runtime"
	"strings"

	"gno.land/p/demo/releases"
	"gno.land/r/gov/dao"
)

var (
	cd        = runtime.ChainDomain()
	changelog = releases.NewChangelog("r/gnoland/users")
)

const usersPrefix = "gno.land/r/gnoland/users/"

func init() {
	changelog.NewRelease("v1", "/r/gnoland/users/v1", "[Original PR](https://github.com/gnolang/gno/pull/3166)")
}

func Render(_ string) string {
	return changelog.RenderAsTable(10)
}

func LatestRelease() string {
	return cd + changelog.Latest().URL()
}

// ProposeNewRelease allows a GovDAO proposal to add a release to the changelog
func ProposeNewRelease(newVerPkgPath, note string) dao.ProposalRequest {
	ver := strings.TrimPrefix(newVerPkgPath, usersPrefix)
	if ver == newVerPkgPath || // TrimPrefix returns unchanged newVerPkgPath if !HasPrefix
		strings.Contains(ver, "/") { // if has prefix, has to be first child under
		panic("r/gnoland/users: invalid version pkgpath")
	}

	cb := func(cur realm) error {
		changelog.NewRelease(ver, strings.TrimPrefix(newVerPkgPath, "gno.land"), note)
		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(
		"Propose a new release for gnoland/users realm",
		"",
		e,
	)
}
