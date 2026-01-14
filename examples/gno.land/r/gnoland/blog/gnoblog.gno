package gnoblog

import (
	"chain/runtime"

	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Title:  "Gno.land's blog",
	Prefix: "/r/gnoland/blog:",
}

func AddComment(_ realm, postSlug, comment string) {
	assertIsCommenter()
	assertNotInPause()

	caller := runtime.OriginCaller()
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
