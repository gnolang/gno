package blog

import (
	"strings"
	"testing"

	"gno.land/p/lou/blog"
)

func clearState(t *testing.T) {
	t.Helper()

	myBlog, _ = blog.NewBlog(
		"Lou's Blog",
		admin,
		blog.WithUserResolver(myResolver),
	)
	// admin = address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs")
	// inPause = false
}

func TestPackage(cur realm, t *testing.T) {
	clearState(t)

	testing.SetOriginCaller(admin)
	testing.SetRealm(testing.NewUserRealm(admin))

	// create two posts, list post.
	{
		CreatePost(cross, "slug1", "title1", "body1", "2022-05-20T13:17:22Z", "lou", "tag1 tag2")
		CreatePost(cross, "slug2", "title2", "body2", "2022-05-20T13:17:23Z", "lou", "tag1 tag3")
		got := Render("")
		expected := `# Lou's Blog

[⊞ grid](/r/lou/blog:?mode=grid) | [≔ list](/r/lou/blog:?mode=list) | [⧖ relative](/r/lou/blog:?time=short) | [↕ alphabetical \(A\-Z\)](/r/lou/blog:?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:?order=asc&sort=recent) | [↕ last updated](/r/lou/blog:?order=asc&sort=update) | [past year](/r/lou/blog:?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

<gno-columns>
<gno-columns>
## [title2](/r/lou/blog:posts/slug2)


##### */slug2*


just now

**tags:** ` `tag1` `, ` `tag3` `

**[comments \(0](/r/lou/blog:posts/slug2#comments)**) | ❤️ (0)

[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug2)



---
|||
## [title1](/r/lou/blog:posts/slug1)


##### */slug1*


just now

**tags:** ` `tag1` `, ` `tag2` `

**[comments \(0](/r/lou/blog:posts/slug1#comments)**) | ❤️ (0)

[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug1)



---
|||
</gno-columns>
`
		assertMDEquals(t, got, expected)
	}

	// view post.
	{
		got := Render("p/slug2")
		expected := `# Lou's Blog\n\n[⊞ grid](/r/lou/blog:?mode=grid) | [≔ list](/r/lou/blog:?mode=list) | [⧖ relative](/r/lou/blog:?time=short) | [↕ alphabetical \\(A\\-Z\\)](/r/lou/blog:?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:?order=asc&sort=recent) | [↕ last updated](/r/lou/blog:?order=asc&sort=update) | [past year](/r/lou/blog:?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)\n\n<gno-columns>\n<gno-columns>\n## [title2](/r/lou/blog:posts/slug2)\n\n\n##### */slug2*\n\n\njust now\n\n**tags:** ` `tag1` `, ` `tag3` `\n\n**[comments \\(0](/r/lou/blog:posts/slug2#comments)**) | ❤️ (0)\n\n[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug2)\n\n\n\n---\n|||\n## [title1](/r/lou/blog:posts/slug1)\n\n\n##### */slug1*\n\n\njust now\n\n**tags:** ` `tag1` `, ` `tag2` `\n\n**[comments \\(0](/r/lou/blog:posts/slug1#comments)**) | ❤️ (0)\n\n[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug1)\n\n\n\n---\n|||\n</gno-columns>`
		assertMDEquals(t, got, expected)
	}

	// list by tags.
	{
		got := Render("tags/invalid")
		expected := `404`
		assertMDEquals(t, got, expected)

		got = Render("tags/tag2")
		expected = ``
		assertMDEquals(t, got, expected)
	}

	// add comments.
	{
		AddCommentToPostBySlug(cross, "slug1", "comment1")
		AddCommentToPostBySlug(cross, "slug2", "comment2")
		AddCommentToPostBySlug(cross, "slug1", "comment3")
		AddCommentToPostBySlug(cross, "slug2", "comment4")
		AddCommentToPostBySlug(cross, "slug1", "comment5")
		got := Render("p/slug2")
		expected := ``
		assertMDEquals(t, got, expected)
	}

	// edit post.
	{
		oldTitle := "title2"
		oldDate := "2022-05-20T13:17:23Z"

		UpdatePostBySlug(cur, "slug2", oldTitle, "body2++", oldDate, "manfred", "tag1,tag4")
		got := Render("p/slug2")
		expected := ``
		assertMDEquals(t, got, expected)

		home := Render("")

		if strings.Count(home, oldTitle) != 1 {
			t.Errorf("post not edited properly")
		}
		// Edits work everything except title, slug, and publicationDate
		// Edits to the above will cause duplication on the blog home page
	}
	//
	{ // Test remove functionality
		title := "example title"
		slug := "testSlug1"
		CreatePost(cross, slug, title, "body1", "2022-05-25T13:17:22Z", "lou", "tag1,tag2")

		got := Render("")

		if !strings.Contains(got, title) {
			t.Errorf("post was not added properly")
		}

		postRender := Render("p/" + slug)

		if !strings.Contains(postRender, title) {
			t.Errorf("post not rendered properly")
		}

		DeletePost(cur, slug)
		got = Render("")

		if strings.Contains(got, title) {
			t.Errorf("post was not removed")
		}

		postRender = Render("p/" + slug)

		assertMDEquals(t, postRender, "404")
	}
	//
	//	// TODO: pagination.
	//	// TODO: ?format=...
	//
	// all 404s
	{
		notFoundPaths := []string{
			"p/slug3",
			"p",
			"p/",
			"x/x",
			"t",
			"t/",
			"/",
			"p/slug1/",
		}
		for _, notFoundPath := range notFoundPaths {
			got := Render(notFoundPath)
			expected := "404"
			if got != expected {
				t.Errorf("path %q: expected %q, got %q.", notFoundPath, expected, got)
			}
		}
	}
}

func assertMDEquals(t *testing.T, got, expected string) {
	t.Helper()
	expected = strings.TrimSpace(expected)
	got = strings.TrimSpace(got)
	if expected != got {
		t.Errorf("invalid render output.\nexpected %q.\ngot      %q.", expected, got)
	}
}
