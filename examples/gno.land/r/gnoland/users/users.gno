package users

import (
	"gno.land/p/demo/releases"
)

var changelog = releases.NewChangelog("r/gnoland/users")

func init() {
	changelog.NewRelease("v1", "/r/gnoland/users/v1", "[Original PR](https://github.com/gnolang/gno/pull/3166)")
}

func Render(_ string) string {
	return changelog.RenderAsTable(10)
}
