package blog_test

import (
	"strings"
	"testing"

	"gno.land/p/demo/seqid"
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
	println("1. NO ERR :: in createTestBlog()")
	b, err := blog.NewBlog("Test Blog", "/r/test/blog")
	if err != nil {
		println("1.1 ERR :: new blog error")
	}
	// urequire.NoError(nil, err, "create blog")
	b.SetUserResolver(testResolver)

	println("1.1 NO ERR :: no user resolver set")
	post, err := blog.NewPost(
		"post-1",
		"First Post",
		"This is the first post body.",
		"2025-06-01T12:00:00Z",
		"g1user1",
		[]string{"g1user2", "g1user3"},
		[]string{"test", "blog"},
	)
	if err != nil {
		println("1.2 ERR :: new post error")
	}
	println("1.2 NO ERR :: post created")
	// urequire.NoError(nil, err, "create post")

	comm1, err := blog.NewComment("g1user2", "I love this post!")
	// urequire.NoError(nil, err, "create comment")

	comm2, err := blog.NewComment("g1user3", "Thanks!")
	// urequire.NoError(nil, err, "create reply")
	comm1.AddReply(comm2)

	post.AddComment(comm1)
	b.AddPost(post)
	return b
}

func TestBlogRender(t *testing.T) {
	println("0. BEGINNING :: createTestBlog()")
	b := createTestBlog()
	out := b.Render("")
	println("DEBUG: blog.Render output:\n", out)

	check := false
	check = strings.Contains(out, "First Post")
	uassert.True(t, check, "First Post")
	check = strings.Contains(out, "This is the first post body.")
	uassert.True(t, check, "This is the first post body.")
	check = strings.Contains(out, "User One")
	uassert.True(t, check, "User One")
	check = strings.Contains(out, "User Two")
	uassert.True(t, check, "User Two")
	check = strings.Contains(out, "User Three")
	uassert.True(t, check, "User Three")
	check = strings.Contains(out, "I love this post!")
	uassert.True(t, check, "I love this post!")
	check = strings.Contains(out, "Thanks!")
	uassert.True(t, check, "Thanks!")
}

func TestBlogPostCommentLinking(t *testing.T) {
	b := createTestBlog()
	post, _ := b.GetPostBySlug("post-1")
	urequire.NotEqual(t, nil, post, "post not nil")

	uassert.Equal(t, "First Post", post.Title())
	uassert.Equal(t, "User One", post.Publisher())

	comments := post.Comments
	urequire.Equal(t, 1, comments.Size(), "one top-level comment")

	currId := seqid.ID(0)
	comment, _ := comments.Get(currId.Next().String())
	comm := comment.(*blog.Comment)
	uassert.Equal(t, "User Two", comm.Author())
	uassert.Equal(t, "I love this post!", comm.Content())

	replies := comm.Replies
	urequire.Equal(t, 1, replies.Size(), "one reply")
	
	reply, _ := replies.Get(currId.String())
	firstReply := reply.(*blog.Comment)
	uassert.Equal(t, "User Three", firstReply.Author())
	uassert.Equal(t, "Thanks!", firstReply.Content())
}

func TestUserRenderLink(t *testing.T) {
	data, ok := users.ResolveAny("g1ngywvql2ql7t8uzl63w60eqcejkwg4rm4lxdw9")
	if !ok || data == nil {
		t.Skip("user not found in test environment")
	}
	link := data.RenderLink("")
	uassert.True(t, strings.HasPrefix(link, "<a href"), "rendered link")
}
