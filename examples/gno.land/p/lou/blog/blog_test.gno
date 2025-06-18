package blog_test

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/lou/blog"
	"gno.land/r/sys/users"
)

func testResolver(input string) (string, bool) {
	if input == "g1user1" {
		return "User One", true
	} else if input == "g1user2" {
		return "User Two", true
	} else if input == "g1user3" {
		return "User Three", true
	}
	return "", false
}

func createTestBlog() *blog.Blog {
	b, err := blog.NewBlog("Test Blog", "/r/test/blog")
	urequire.NoError(nil, err, "create blog")
	b.SetUserResolver(testResolver)

	post, err := blog.NewPost(
		"post-1",
		"First Post",
		"This is the first post body.",
		"2025-06-01T12:00:00Z",
		"g1user1",
		[]string{"g1user2", "g1user3"},
		[]string{"test", "blog"},
	)
	urequire.NoError(nil, err, "create post")

	comm1, err := blog.NewComment("g1user2", "I love this post!")
	urequire.NoError(nil, err, "create comment")

	comm2, err := blog.NewComment("g1user3", "Thanks!")
	urequire.NoError(nil, err, "create reply")
	comm1.AddReply(comm2)

	post.AddComment(comm1)
	b.AddPost(post)
	return b
}

// func TestBlogRender(t *testing.T) {
// 	b := createTestBlog()
// 	out := b.Render("/r/test/blog")

// 	uassert.StringContains(t, out, "First Post")
// 	uassert.StringContains(t, out, "This is the first post body.")
// 	uassert.StringContains(t, out, "User One")
// 	uassert.StringContains(t, out, "User Two")
// 	uassert.StringContains(t, out, "User Three")
// 	uassert.StringContains(t, out, "I love this post!")
// 	uassert.StringContains(t, out, "Thanks!")
// }

func TestBlogPostCommentLinking(t *testing.T) {
	b := createTestBlog()
	post, _ := b.GetPostBySlug("post-1")
	urequire.NotEqual(t, nil, post, "post not nil")

	uassert.Equal(t, "First Post", post.Title())
	uassert.Equal(t, "User One", post.Author())

	comments := post.Comments()
	urequire.Equal(t, 1, len(comments), "one top-level comment")

	comm := comments[0]
	uassert.Equal(t, "User Two", comm.Author())
	uassert.Equal(t, "I love this post!", comm.Body())

	replies := comm.Replies()
	urequire.Equal(t, 1, len(replies), "one reply")
	uassert.Equal(t, "User Three", replies[0].Author())
	uassert.Equal(t, "Thanks!", replies[0].Body())
}

func TestUserRenderLink(t *testing.T) {
	data, ok := users.ResolveAny("g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9")
	if !ok || data == nil {
		t.Skip("user not found in test environment")
	}
	link := data.RenderLink("")
	uassert.True(t, strings.HasPrefix(link, "<a href"), "rendered link")
}
