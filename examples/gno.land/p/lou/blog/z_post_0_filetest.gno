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
// # test blog
// 
// [⊞ grid](/r/test:?mode=grid) | [☰ list](/r/test:?mode=list) | [⏱ relative](/r/test:?time=short) | [↕ alphabetical \(A\-Z\)](/r/test:?order=asc&sort=alpha) | [↕ recent](/r/test:?order=asc&sort=recent) | [↕ last updated](/r/test:?order=asc&sort=update) | [past year](/r/test:?end=2009-02-13&start=2008-02-13), [this year](/r/test:?end=2009-02-13&start=2009-01-01), [last 30 days](/r/test:?end=2009-02-13&start=2009-01-14) | [🗘 reset](/r/test)
// 
// No posts found.