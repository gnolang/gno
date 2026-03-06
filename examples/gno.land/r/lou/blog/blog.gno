package blog

import (
	"gno.land/p/lou/blog"

	"gno.land/r/leon/hor"
	"gno.land/r/sys/users"
)

var (
	myBlog *blog.Blog
	admin  = address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs")
)

func init() {
	myBlog, _ = blog.NewBlog(
		"Lou's Blog",
		admin,
		blog.WithUserResolver(myResolver),
	)
	hor.Register(cross, "Lou's blog realm", "")
}

func Render(path string) string {
	out := ""
	out += myBlog.Render(path)
	return out
}

func myResolver(input string) (string, bool) {
	data, ok := users.ResolveAny(input)
	if !ok || data == nil {
		return "", false
	}
	return data.Name(), true
}
