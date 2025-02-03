package users

import (
	"strconv"

	"gno.land/p/demo/releases"
)

var changelog = releases.NewChangelog("r/gnoland/users")

func init() {
	for i := 1; i < 11; i++ {
		changelog.NewRelease("v"+strconv.Itoa(i), "/r/gnoland/users/v1", "Version 1 of the gno.land user registration system.")
	}
}

func Render(_ string) string {
	return changelog.RenderAsTable(10)
}
