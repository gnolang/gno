package gnoblog

import (
	"std"

	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Title:  "Gnoland's Blog",
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
	return b.Render(path)
}

func RenderLastPostsWidget(limit int) string {
	return b.RenderLastPostsWidget(limit)
}

func PostExists(slug string) bool {
	if b.GetPost(slug) == nil {
		return false
	}
	return true
}
