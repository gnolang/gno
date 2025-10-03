package main

import (
	"std"
	"testing"

	"gno.land/p/lou/blog"
)

func main() {
	user := std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs")
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/postblog"))

	blogInstance, _ := blog.NewBlog("Devlog", user)
	post, _ := blog.NewPost("first-post", "Hello world", "the body", "2024-03-01T10:00:00Z", user.String(), []string{"g1abcde12345"},
		[]string{"ai", "technology", "future"})
	blogInstance.AddPost(post)
	println(blogInstance.Render("/first-post"))
}

// Output:
//
