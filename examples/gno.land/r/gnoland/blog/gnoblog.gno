package gnoblog

import "gno.land/p/blog"

var b = &blog.Blog{
	Prefix: "r/gnoland/blog:",
}

func Render(path string) string {
	output := "# Gnoland's Blog\n\n"
	output += b.Render(path)
	return output
}
