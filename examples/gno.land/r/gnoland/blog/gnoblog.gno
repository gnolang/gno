package gnoblog

import (
	"std"
	"strings"

	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Prefix: "r/gnoland/blog:",
}

func ModAddPost(slug, title, body, tags string) {
	assertIsModerator()

	caller := std.GetOrigCaller()
	tagList := strings.Split(tags, ",")
	err := b.NewPost(caller, slug, title, body, tagList)
	checkErr(err)
}

func ModEditPost(slug, title, body, tags string) {
	assertIsModerator()

	tagList := strings.Split(tags, ",")
	err := b.GetPost(slug).Update(title, body, tagList)
	checkErr(err)
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
