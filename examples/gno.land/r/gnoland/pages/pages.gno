package pages

import (
	"gno.land/p/demo/blog"
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
