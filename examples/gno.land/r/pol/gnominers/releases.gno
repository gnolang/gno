package gnominers

import (
	"std"
	"strings"

	"gno.land/p/demo/releases"
	"gno.land/p/moul/md"
	"gno.land/p/pol/dao"
	"gno.land/p/pol/dao/simple"
)

var (
	changelog = releases.NewChangelog("r/pol/gnominers")
	herePath = std.CurrentRealm().PkgPath()
	here = herePath[len(std.ChainDomain()):]
)

func init() {
	changelog.NewRelease("v1", here + "/v1", md.Link("Original PR", "https://github.com/gnolang/gno/pull/4071"))
}

func Render(_ string) string {
	return changelog.RenderAsTable(10) + "\n\n" + here
}

func LatestRelease() string {
	return changelog.Latest().URL()
}

func ProposeNewRelease(path, note string) *simple.SimpleProposal {
	version := strings.TrimPrefix(path, here + "/")
	return simple.NewSimpleProposal(dao.Member(""), "New Release", "new released proposed", func() error {
		changelog.NewRelease(version, strings.TrimPrefix(path, std.ChainDomain()), note)
		return nil
	})
}
