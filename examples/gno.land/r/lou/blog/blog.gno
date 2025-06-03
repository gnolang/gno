package blog

import (
	"std"

	"gno.land/p/lou/blog"
	"gno.land/p/moul/md"
)

var myBlog = blog.NewBlog(
	"Blog",
	"/r/lou/blog",
)

func Render(path string) string {
	out := md.H1("Blog")
	// just to test for now
	postone, err := blog.NewPost("slug", "Titlehere", "Bodyis here now", "2004-10-28T00:00:00Z", []string{"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"}, []string{"me", "myself", "i"})
	postone.AddComment(blog.NewComment(std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"), "This is a comment"))
	posttwo, err := blog.NewPost("slug2", "Titlehere2", "Bodyis here now2", "2004-10-28T00:00:00Z", []string{"lou"}, []string{"yo", "myself", "i"})
	if err != nil {
		out += "Error creating post: " + err.Error()
	} else {
		myBlog.AddPostToBlog(postone)
		myBlog.AddPostToBlog(postone)
		myBlog.AddPostToBlog(postone)
		myBlog.AddPostToBlog(posttwo)
		myBlog.AddPostToBlog(postone)
		myBlog.AddPostToBlog(postone)
		myBlog.AddPostToBlog(posttwo)
		myBlog.AddPostToBlog(posttwo)
		myBlog.AddPostToBlog(posttwo)
		myBlog.AddPostToBlog(posttwo)
		myBlog.AddPostToBlog(posttwo)
	}

	out += myBlog.Render(path)
	return out
}
