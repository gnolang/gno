package blog

import (
	"std"

	"gno.land/p/lou/blog"

	"gno.land/r/sys/users"
)

var (
	myBlog *blog.Blog
)

func init() {
	var err error
	myBlog, err = blog.NewBlog(
		"Lou's Blog",
		std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"),
		blog.WithUserResolver(myResolver),
	)
	if err != nil {
		panic(err)
	}
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
