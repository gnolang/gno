package blog

import (
	"std"

	"gno.land/p/lou/blog"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
)

var myBlog = blog.NewBlog(
	"Blog",
	"r/gnoland/blog:",
)

func ToggleGrid() {
	crossing()
	myBlog.ToggleGridMode()
}

func ToggleList() {
	crossing()
	myBlog.ToggleListMode()
}

func Render(path string) string {
	out := md.H1("Blog")
	out += md.Link("(toggle grid mode", txlink.Call("ToggleGrid")) + " | "
	out += md.Link("toggle list mode)", txlink.Call("ToggleList")) + "\n\n"

	postone, err := blog.NewPost("slug", "Titlehere", "Bodyis here now", "2004-10-28T00:00:00Z", []string{std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs").String()}, []string{"me", "myself", "i"})
	if err != nil {
		out += "Error creating post: " + err.Error()
	} else {
		myBlog.AddPostToBlog(postone)
	}
	out += myBlog.Render(path)
	return out
}
