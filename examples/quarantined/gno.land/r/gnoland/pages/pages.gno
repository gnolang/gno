package pages

import (
	"gno.land/p/gnoland/blog/v0"
)

// TODO: switch from p/blog to p/pages

var b = &blog.Blog{
	Title:        "Gnoland's Pages",
	Prefix:       "/r/gnoland/pages:",
	NoBreadcrumb: true,
}

func Render(path string) string {
	return b.Render(path)
}
