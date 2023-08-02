package gnopages

import (
	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Title:  "Gnoland's Pages",
	Prefix: "/r/gnoland/pages:",
}

func init() {
	_ = b.NewPost("", "gor", "Game of Realms", "Lorem Ipsum", nil)
	_ = b.NewPost("", "events", "Events", "Lorem Ipsum", nil)
	_ = b.NewPost("", "tokenomics", "Tokenomics", "Lorem Ipsum", nil)
	_ = b.NewPost("", "start", "Getting Started", "Lorem Ipsum", nil)
}

func Render(path string) string {
	return b.Render(path)
}
