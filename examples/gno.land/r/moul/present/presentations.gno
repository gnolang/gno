package present

import (
	"gno.land/p/demo/blog"
)

// TODO: switch from p/blog to p/present

var b = &blog.Blog{
	Title:        "Manfred's Presentations",
	Prefix:       "/r/moul/present:",
	NoBreadcrumb: true,
}

func Render(path string) string {
	return b.Render(path)
}
