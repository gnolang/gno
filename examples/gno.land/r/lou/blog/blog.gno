package blog

import (
	"std"

	"gno.land/p/lou/blog"
)

var (
	myBlog *blog.Blog
)

func init() {
	var err error
	myBlog, err = blog.NewBlog("Lou's Blog", "/r/lou/blog", std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"))
	if err != nil {
		panic(err)
	}
}

func Render(path string) string {
	out := ""
	out += myBlog.Render(path)
	return out
}
