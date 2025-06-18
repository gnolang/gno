package blog_test

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/lou/blog"
)

func testUserResolver(input string) (string, bool) {
	switch input {
	case "g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs":
		return "AuthorOne", true
	case "g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n":
		return "CommenterOne", true
	case "g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae":
		return "ReplyUser", true
	default:
		return "", false
	}
}

func TestBlogRenderBasic(t *testing.T) {
	blogTest, err := blog.NewBlog("Test Blog", "/r/test/blog")
	urequire.NoError(t, err, "creating blog")
	blogTest.SetUserResolver(testUserResolver)

	post, err := blog.NewPost(
		"sample-post",
		"My First Post",
		"Hello world blog content.",
		"2025-06-18T10:00:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n"},
		[]string{"intro", "welcome"},
	)
	urequire.NoError(t, err, "creating post")

	comment, err := blog.NewComment("g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n", "Nice Post!")
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
	check = strings.Contains(out, "CommenterOne")
	uassert.True(t, check, "CommenterOne")
	check = strings.Contains(out, "User Two")
	uassert.False(t, check, "User Two")
	check = strings.Contains(out, "I love this post!")
	uassert.False(t, check, "I love this post!")
	check = strings.Contains(out, "Nice Post!")
	uassert.True(t, check, "Nice Post!")
}

func TestCommentThreadRender(t *testing.T) {
	blogTest, err := blog.NewBlog("Thread Test", "/r/test/blog")
	urequire.NoError(t, err, "creating blog")
	blogTest.SetUserResolver(testUserResolver)

	post, err := blog.NewPost(
		"threaded-post",
		"Thread Test",
		"Checking comment threads.",
		"2025-06-18T11:00:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n"},
		[]string{"threads"},
	)
	urequire.NoError(t, err, "creating post")

	// Parent comment
	parent, err := blog.NewComment("g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n", "Needs more detail.")
	urequire.NoError(t, err, "creating parent comment")

	// Reply
	reply, err := blog.NewComment("g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae", "Agreed.")
	urequire.NoError(t, err, "creating reply")
	parent.AddReply(reply)

	err = post.AddComment(parent)
	println(parent.Author())
	urequire.NoError(t, err, "adding comment to post ("+err.Error()+")")
	err = blogTest.AddPost(post)
	urequire.NoError(t, err, "adding post to blog")

	out := blogTest.Render("posts/threaded-post")
	check := false
	check = strings.Contains(out, "AuthorOne")
	uassert.True(t, check, "AuthorOne")
	check = strings.Contains(out, "Needs more detail.")
	uassert.True(t, check, "Needs more detail.")
	check = strings.Contains(out, "Agreed.")
	uassert.True(t, check, "Agreed.")
	check = strings.Contains(out, "CommenterOne")
	uassert.True(t, check, "CommenterOne")
	check = strings.Contains(out, "ReplyUser")
	uassert.True(t, check, "ReplyUser")
}
