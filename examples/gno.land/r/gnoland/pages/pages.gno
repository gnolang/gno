package gnopages

import (
	"std"

	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Title:  "Gnoland's Pages",
	Prefix: "/r/gnoland/pages:",
}

func init(){
	_ = b.NewPost("", "gor", "Game of Realms", "Lorem Ipsum", nil)
	_ = b.NewPost("", "events", "Events", "Lorem Ipsum", nil)
}

func Render(path string) string {
	return b.Render(path)
}
