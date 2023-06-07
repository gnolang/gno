package gnopages

import (
	"std"

	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Title:  "Gnoland's Pages",
	Prefix: "/r/gnoland/pages:",
}

func Render(path string) string {
	return b.Render(path)
}
