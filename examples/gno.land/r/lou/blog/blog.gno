package blog

import (
	"gno.land/p/lou/blog"
)

var (
	myBlog *blog.Blog
)

func init() {
	var err error
	myBlog, err = blog.NewBlog("Lou's Blog", "/r/lou/blog")
	if err != nil {
		panic(err)
	}
}

func Render(path string) string {
	out := ""
	out += myBlog.Render(path)
	return out
}
