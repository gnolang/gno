package userprofile

import (
	_ "gno.land/r/demo/profile"
	_ "gno.land/r/gnoland/users"
	_ "gno.land/r/sys/users"
)

func Render(path string) string {
	return "TEST"
}
