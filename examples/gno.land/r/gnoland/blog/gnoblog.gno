package gnoblog

import (
	"std"

	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Prefix: "/r/gnoland/blog:",
}

func AddComment(postSlug, comment string) {
	assertIsCommenter()
	assertNotInPause()

	caller := std.GetOrigCaller()
	err := b.GetPost(postSlug).AddComment(caller, comment)
	checkErr(err)
}

func Render(path string) string {
	output := "# Gnoland's Blog\n\n"
	output += b.Render(path)
	return output
}
