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
	out := md.H1("Blog") + md.HorizontalRule()

	out += md.Link("toggle grid mode", txlink.Call("ToggleGrid")) + "\n\n"
	out += md.Link("toggle list mode", txlink.Call("ToggleList")) + "\n\n"

	myBlog.AddPostToBlog(blog.NewPost("slug", "Titlehere", "Bodyis here now", "28 Oct 2004", []string{std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs").String()}, []string{"me", "myself", "i"}))
	myBlog.AddPostToBlog(blog.NewPost("slug", "Titlehere", "Bodyis here now", "28 Oct 2004", []string{std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs").String()}, []string{"me", "myself", "i"}))
	myBlog.AddPostToBlog(blog.NewPost("slug", "Titlehere", "Bodyis here now", "28 Oct 2004", []string{std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs").String()}, []string{"me", "myself", "i"}))
	myBlog.AddPostToBlog(blog.NewPost("slug", "Titlehere", "Bodyis here now", "28 Oct 2004", []string{std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs").String()}, []string{"me", "myself", "i"}))
	myBlog.AddPostToBlog(blog.NewPost("slug", "Titlehere", "Bodyis here now", "28 Oct 2004", []string{std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs").String()}, []string{"me", "myself", "i"}))

	out += myBlog.RenderPosts()
	return out
}
