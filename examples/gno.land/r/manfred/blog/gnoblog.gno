package gnoblog

import (
	"std"

	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Title:  "Manfred's Blog",
	Prefix: "/r/manfred/blog:",
}

func Render(path string) string {
	return b.Render(path)
}
