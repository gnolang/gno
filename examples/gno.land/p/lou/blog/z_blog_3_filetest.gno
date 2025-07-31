package main

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/lou/blog"
)

// Simulated user resolver function
func testUserResolver(input string) (string, bool) {
	switch input {
	case "g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs":
		return "authorone", true
	case "g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n":
		return "commenter1", true
	case "g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae":
		return "replyuser", true
	default:
		return "", false
	}
}

func main() {
	// User for test
	user := std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs")

	// Realm setup
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	// Create blog
	b, err := blog.NewBlog("Test Blog", user)
	if err != nil {
		panic("blog create: " + err.Error())
	}

	b.Authorizable.AddToAuthList(user)
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	b.SetUserResolver(testUserResolver)

	// Create a post
	post, err := blog.NewPost(
		"sample-post",
		"My First Post",
		"Hello world blog content.",
		"2025-06-18T10:00:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n"},
		[]string{"intro", "welcome"},
	)
	if err != nil {
		panic("post create: " + err.Error())
	}

	comment, err := blog.NewComment("g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n", "Nice Post!")
	if err != nil {
		panic("comment create: " + err.Error())
	}

	testing.SetRealm(std.NewUserRealm(user))
	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	post.AddComment(comment)

	testing.SetRealm(std.NewUserRealm(user))
	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	b.AddPost(post)

	// Render list
	out := b.Render("")
	if !strings.Contains(out, "First Post") {
		panic("missing post in render list")
	}

	// Render individual post
	out = b.Render("posts/sample-post")
	if !strings.Contains(out, "Nice Post!") {
		panic("missing comment")
	}
	if !strings.Contains(out, "commenter1") {
		panic("missing commenter1 username")
	}

	println("✅ blog render basic passed")
}

// Output:
//
