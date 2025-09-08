package blog

// Blog is created with a title and path.

// Blog rejects invalid input (e.g. empty title/path).

// Blog renders with no posts (empty blog).

// Blog renders correctly with one post and no comments.

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func clearState(t *testing.T) {
	t.Helper()

	//	b = &blog.Blog{
	//		Title:  "gno.land's blog",
	//		Prefix: "/r/gnoland/blog:",
	//	}
	//
	// adminAddr = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	// inPause = false
	// moderatorList = avl.NewTree()
	// commenterList = avl.NewTree()
}

func TestBlogRenderBasicB(t *testing.T) {
	clearState(t)

	var (
		author    = std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs")
		commenter = std.Address("g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n")
	)

	testing.SetOriginCaller(author)
	testing.SetRealm(std.NewUserRealm(author))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/lou/blog"))

	blogTest, err := NewBlog("Test Blog", author)
	urequire.NoError(t, err)

	testing.SetRealm(std.NewUserRealm(author)) // re-enter caller realm
	testing.SetRealm(std.NewCodeRealm("gno.land/r/lou/blog"))
	blogTest.SetUserResolver(testUserResolver)

	post, err := NewPost(
		"sample-post",
		"My First Post",
		"Hello world blog content.",
		"2025-06-18T10:00:00Z",
		author.String(),
		[]string{commenter.String()},
		[]string{"intro", "welcome"},
	)
	urequire.NoError(t, err)

	comment, err := NewComment(commenter.String(), "Nice Post!")
	urequire.NoError(t, err)
	post.AddComment(comment)

	blogTest.AddPost(post)

	out := blogTest.Render("")
	uassert.True(t, strings.Contains(out, "First Post"), "should contain title")

	out = blogTest.Render("posts/sample-post")
	uassert.True(t, strings.Contains(out, "Nice Post!"), "should contain comment")
	uassert.True(t, strings.Contains(out, "commenter1"), "should resolve username")
	uassert.False(t, strings.Contains(out, "I love this post!"), "should not include unrelated content")
}
