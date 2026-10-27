package blog

import "gno.land/p/gnoland/blog/v0"

var b = &blog.Blog{
	Title:  "Gno.land's blog",
	Prefix: "/r/gnoland/blog:",
}

func AddComment(cur realm, postSlug, comment string) {
	caller := cur.Previous().Address()
	assertIsCommenter(caller)
	assertNotInPause()

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
