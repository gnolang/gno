package main

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/lou/blog"
)

// TO FIX
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
	user := std.Address("g1testuser")
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetOriginCaller(user)
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	b, err := blog.NewBlog("Thread Test", user)
	if err != nil {
		panic("blog create: " + err.Error())
	}

	b.Authorizable.AddToAuthList(user)
	b.SetUserResolver(testUserResolver)

	post, err := blog.NewPost(
		"threaded-post",
		"Thread Test",
		"Checking comment threads.",
		"2025-06-18T11:00:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n"},
		[]string{"threads"},
	)
	if err != nil {
		panic("post create: " + err.Error())
	}

	parent, err := blog.NewComment("g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n", "Needs more detail.")
	if err != nil {
		panic("parent comment: " + err.Error())
	}
	post.AddComment(parent)

	reply, err := blog.NewComment("g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae", "Agreed.")
	if err != nil {
		panic("reply: " + err.Error())
	}
	post.AddReply(parent.ID(), reply)

	b.AddPost(post)

	out := b.Render("posts/threaded-post")
	if !strings.Contains(out, "Checking comment threads.") ||
		!strings.Contains(out, "Needs more detail.") ||
		!strings.Contains(out, "Agreed.") ||
		!strings.Contains(out, "commenter1") ||
		!strings.Contains(out, "replyuser") {
		panic("thread rendering failed")
	}

	println("✅ thread render passed")
}

// Output:
// # test blog
//
// [⊞ grid](/r/test:?mode=grid) | [☰ list](/r/test:?mode=list) | [⏱ relative](/r/test:?time=short) | [↕ alphabetical \(A\-Z\)](/r/test:?order=asc&sort=alpha) | [↕ recent](/r/test:?order=asc&sort=recent) | [↕ last updated](/r/test:?order=asc&sort=update) | [past year](/r/test:?end=2009-02-13&start=2008-02-13), [this year](/r/test:?end=2009-02-13&start=2009-01-01), [last 30 days](/r/test:?end=2009-02-13&start=2009-01-14) | [🗘 reset](/r/test)
//
// No posts found.
