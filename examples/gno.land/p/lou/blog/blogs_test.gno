package blog

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

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

func TestBlogRenderBasic(t *testing.T) {
	blogTest, err := NewBlog("Test Blog", std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"))
	urequire.NoError(t, err, "creating blog")
	blogTest.SetUserResolver(testUserResolver)

	post, err := NewPost(
		"sample-post",
		"My First Post",
		"Hello world blog content.",
		"2025-06-18T10:00:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n"},
		[]string{"intro", "welcome"},
	)
	urequire.NoError(t, err, "creating post")

	comment, err := NewComment("g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n", "Nice Post!")
	urequire.NoError(t, err, "creating comment")
	post.AddComment(comment)

	blogTest.AddPost(post)
	check := false

	out := blogTest.Render("")
	check = strings.Contains(out, "First Post")
	uassert.True(t, check, "First Post")

	out = blogTest.Render("posts/sample-post")
	check = strings.Contains(out, "This is the first post body.")
	uassert.False(t, check, "This is the first post body.")
	check = strings.Contains(out, "User One")
	uassert.False(t, check, "User One")
	check = strings.Contains(out, "commenter1")
	uassert.True(t, check, "commenter1")
	check = strings.Contains(out, "user2")
	uassert.False(t, check, "user2")
	check = strings.Contains(out, "I love this post!")
	uassert.False(t, check, "I love this post!")
	check = strings.Contains(out, "Nice Post!")
	uassert.True(t, check, "Nice Post!")
}

func TestCommentThreadRender(t *testing.T) {
	blogTest, err := NewBlog("Thread Test", std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"))
	urequire.NoError(t, err, "creating blog")
	blogTest.SetUserResolver(testUserResolver)

	post, err := NewPost(
		"threaded-post",
		"Thread Test",
		"Checking comment threads.",
		"2025-06-18T11:00:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n"},
		[]string{"threads"},
	)
	urequire.NoError(t, err, "creating post")

	parent, err := NewComment("g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n", "Needs more detail.")
	urequire.NoError(t, err, "creating parent comment")
	post.AddComment(parent)

	reply, err := NewComment("g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae", "Agreed.") // TO FIX
	urequire.NoError(t, err, "creating reply")
	post.AddReply(parent.ID(), reply)

	blogTest.AddPost(post)
	check := false

	out := blogTest.Render("posts/threaded-post")
	check = strings.Contains(out, "Checking comment threads.")
	uassert.True(t, check, "Checking comment threads.")
	check = strings.Contains(out, "authorone")
	uassert.True(t, check, "authorone")
	check = strings.Contains(out, "Needs more detail.")
	uassert.True(t, check, "Needs more detail.")
	check = strings.Contains(out, "Agreed.")
	uassert.True(t, check, "Agreed.")
	check = strings.Contains(out, "commenter1")
	uassert.True(t, check, "commenter1")
	check = strings.Contains(out, "replyuser")
	uassert.True(t, check, "replyuser")
}
