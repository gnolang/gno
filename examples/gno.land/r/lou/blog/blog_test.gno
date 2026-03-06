package blog

import (
	"strings"
	"testing"

	"gno.land/p/lou/blog"
)

func clearState(t *testing.T) {
	t.Helper()
	var err error

	admin = address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs")
	myBlog, err = blog.NewBlog(
		"Lou's Blog",
		admin,
		blog.WithUserResolver(myResolver),
	)
	if err != nil {
		t.Errorf(err.Error())
	}
	// inPause = false
}

func TestPackage(cur realm, t *testing.T) {
	clearState(t)

	testing.SetOriginCaller(admin)
	testing.SetRealm(testing.NewUserRealm(admin))

	{
		println("Creating posts and rendering previews...")
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

**tags:** ` + "`" + `tag1` + "`" + `, ` + "`" + `tag3` + "`" + `

**[comments \(0](/r/lou/blog:posts/slug2#comments)**) | ❤️ (0)

[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug2)



---
|||
## [title1](/r/lou/blog:posts/slug1)


##### */slug1*


just now

**tags:** ` + "`" + `tag1` + "`" + `, ` + "`" + `tag2` + "`" + `

**[comments \(0](/r/lou/blog:posts/slug1#comments)**) | ❤️ (0)

[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug1)



---
|||
</gno-columns>`
		assertMDEquals(t, got, expected)
	}

	{
		println("Viewing one of the post...")
		got := Render("posts/slug2")
		expected := `# title2


*Author(s):* [lou](/r/lou/blog:authors/lou)


body2

---
**Created on:** May 20, 2022 at 1:17 PM

**Published on:** February 13, 2009 at 11:31 PM

**Publisher:** [g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs](/u/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs)

**Tags:** [#tag1](/r/lou/blog:tags/tag1), [#tag3](/r/lou/blog:tags/tag3)


❤️ 0

---
### Comments (0)


No comments yet.`
		assertMDEquals(t, got, expected)
	}

	{
		println("Listing of tags...")
		println("Listing")
		got := Render("authors")
		expected := `## [/r/lou/blog](/r/lou/blog)/authors
[↕ alphabetical \(A\-Z\)](/r/lou/blog:authors?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:authors?order=asc&sort=recent) | [↕ most common](/r/lou/blog:authors?order=asc&sort=common) | [past year](/r/lou/blog:authors?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:authors?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:authors?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

### [lou](/r/lou/blog:authors/lou) (2)`
		assertMDEquals(t, got, expected)

		println("Invalid tag")
		got = Render("tags/invalid")
		expected = `## [/r/lou/blog](/r/lou/blog)/[tags](/r/lou/blog:tags)/invalid
[⊞ grid](/r/lou/blog:tags/invalid?mode=grid) | [≔ list](/r/lou/blog:tags/invalid?mode=list) | [⧖ relative](/r/lou/blog:tags/invalid?time=short) | [↕ alphabetical \(A\-Z\)](/r/lou/blog:tags/invalid?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:tags/invalid?order=asc&sort=recent) | [↕ last updated](/r/lou/blog:tags/invalid?order=asc&sort=update) | [past year](/r/lou/blog:tags/invalid?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:tags/invalid?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:tags/invalid?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

No posts found for this tag.<gno-columns>`
		assertMDEquals(t, got, expected)

		println("Valid tag")
		got = Render("tags/tag2")
		expected = `## [/r/lou/blog](/r/lou/blog)/[tags](/r/lou/blog:tags)/tag2
[⊞ grid](/r/lou/blog:tags/tag2?mode=grid) | [≔ list](/r/lou/blog:tags/tag2?mode=list) | [⧖ relative](/r/lou/blog:tags/tag2?time=short) | [↕ alphabetical \(A\-Z\)](/r/lou/blog:tags/tag2?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:tags/tag2?order=asc&sort=recent) | [↕ last updated](/r/lou/blog:tags/tag2?order=asc&sort=update) | [past year](/r/lou/blog:tags/tag2?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:tags/tag2?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:tags/tag2?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

<gno-columns>
<gno-columns>
## [title1](/r/lou/blog:posts/slug1)


##### */slug1*


just now

**author(s):** lou

**tags:** ` + "`tag1`, `tag2`" + `

**[comments \(0](/r/lou/blog:posts/slug1#comments)**) | ❤️ (0)

[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug1)



---

</gno-columns>`
		assertMDEquals(t, got, expected)
	}

	{
		println("Listing of authors...")
		println("Listing")
		got := Render("authors")
		expected := `## [/r/lou/blog](/r/lou/blog)/authors
[↕ alphabetical \(A\-Z\)](/r/lou/blog:authors?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:authors?order=asc&sort=recent) | [↕ most common](/r/lou/blog:authors?order=asc&sort=common) | [past year](/r/lou/blog:authors?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:authors?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:authors?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

### [lou](/r/lou/blog:authors/lou) (2)`
		assertMDEquals(t, got, expected)

		println("Invalid author")
		got = Render("authors/invalid")
		expected = `## [/r/lou/blog](/r/lou/blog)/[authors](/r/lou/blog:authors)/invalid
#### [@invalid](/u/invalid)'s profile
[⊞ grid](/r/lou/blog:authors/invalid?mode=grid) | [≔ list](/r/lou/blog:authors/invalid?mode=list) | [⧖ relative](/r/lou/blog:authors/invalid?time=short) | [↕ alphabetical \(A\-Z\)](/r/lou/blog:authors/invalid?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:authors/invalid?order=asc&sort=recent) | [↕ last updated](/r/lou/blog:authors/invalid?order=asc&sort=update) | [past year](/r/lou/blog:authors/invalid?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:authors/invalid?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:authors/invalid?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

No posts found for this author.<gno-columns>`
		assertMDEquals(t, got, expected)

		println("Valid author")
		got = Render("authors/lou")
		expected = `## [/r/lou/blog](/r/lou/blog)/[authors](/r/lou/blog:authors)/lou
#### [@lou](/u/lou)'s profile
[⊞ grid](/r/lou/blog:authors/lou?mode=grid) | [≔ list](/r/lou/blog:authors/lou?mode=list) | [⧖ relative](/r/lou/blog:authors/lou?time=short) | [↕ alphabetical \(A\-Z\)](/r/lou/blog:authors/lou?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:authors/lou?order=asc&sort=recent) | [↕ last updated](/r/lou/blog:authors/lou?order=asc&sort=update) | [past year](/r/lou/blog:authors/lou?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:authors/lou?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:authors/lou?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

<gno-columns>
<gno-columns>
## [title2](/r/lou/blog:posts/slug2)


##### */slug2*


just now

**author(s):** lou

**tags:** ` + "`tag1`, `tag3`" + `

**[comments \(0](/r/lou/blog:posts/slug2#comments)**) | ❤️ (0)

[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug2)



---

## [title1](/r/lou/blog:posts/slug1)


##### */slug1*


just now

**author(s):** lou

**tags:** ` + "`tag1`, `tag2`" + `

**[comments \(0](/r/lou/blog:posts/slug1#comments)**) | ❤️ (0)

[Like](/r/lou/blog$help&func=LikePostBySlug&slug=slug1)



---

</gno-columns>`
		assertMDEquals(t, got, expected)
	}

	{
		println("Adding comments on a post...")
		AddCommentToPostBySlug(cross, "slug1", "comment1")
		AddCommentToPostBySlug(cross, "slug2", "comment2")
		AddCommentToPostBySlug(cross, "slug1", "comment3")
		AddCommentToPostBySlug(cross, "slug2", "comment4")
		AddCommentToPostBySlug(cross, "slug1", "comment5")
		got := Render("posts/slug2")
		expected := `# title2


[2 Comment\(s\)](#comments)

*Author(s):* [lou](/r/lou/blog:authors/lou)


body2

---
**Created on:** May 20, 2022 at 1:17 PM

**Published on:** February 13, 2009 at 11:31 PM

**Publisher:** [g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs](/u/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs)

**Tags:** [#tag1](/r/lou/blog:tags/tag1), [#tag3](/r/lou/blog:tags/tag3)


❤️ 0

---
### Comments (2)


**[@g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs)** *just now*

comment4
❤️ 0 



**[@g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs)** *just now*

comment2
❤️ 0`
		assertMDEquals(t, got, expected)
	}

	{
		println("Listing of commenters...")
		println("Listing invalid")
		got := Render("commenters")
		expected := `Commenter slug is required.`
		assertMDEquals(t, got, expected)

		println("Invalid commenter")
		got = Render("commenters/invalid")
		expected = `## [/r/lou/blog](/r/lou/blog)/commenters/invalid
#### [@invalid](/u/invalid)'s profile
[⊞ grid](/r/lou/blog:commenters/invalid?mode=grid) | [≔ list](/r/lou/blog:commenters/invalid?mode=list) | [⧖ relative](/r/lou/blog:commenters/invalid?time=short) | [↕ alphabetical \(A\-Z\)](/r/lou/blog:commenters/invalid?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:commenters/invalid?order=asc&sort=recent) | [↕ last updated](/r/lou/blog:commenters/invalid?order=asc&sort=update) | [past year](/r/lou/blog:commenters/invalid?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:commenters/invalid?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:commenters/invalid?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

No posts found for this commenter.<gno-columns>`
		assertMDEquals(t, got, expected)

		println("Valid commenter")
		got = Render("commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs")
		expected = `## [/r/lou/blog](/r/lou/blog)/commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs
#### [@g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs](/u/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs)'s profile
[⊞ grid](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?mode=grid) | [≔ list](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?mode=list) | [⧖ relative](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?time=short) | [↕ alphabetical \(A\-Z\)](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?order=asc&sort=alpha) | [↕ recent](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?order=asc&sort=recent) | [↕ last updated](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?order=asc&sort=update) | [past year](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?end=2009-02-13&start=2008-02-13), [this year](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?end=2009-02-13&start=2009-01-01), [last 30 days](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs?end=2009-02-13&start=2009-01-14) | [⟳ reset](/r/lou/blog)

<gno-columns>
<gno-columns>
#### **[@g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs)**


comment1

*just now*

in [title1](/r/lou/blog:posts/slug1)

---
#### **[@g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs)**


comment3

*just now*

in [title1](/r/lou/blog:posts/slug1)

---
#### **[@g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs](/r/lou/blog:commenters/g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs)**


comment5

*just now*

in [title1](/r/lou/blog:posts/slug1)

---
</gno-columns>
**1** | [2](?page=2)`
		assertMDEquals(t, got, expected)
	}

	{
		println("Invalid paths...")
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
			} else {
				println("PASSED")
			}
		}
	}

	// {
	// 	println("Updating a post...")
	// 	newTitle := "title3"
	// 	newDate := "2022-05-20T13:17:23Z"

	// 	CreatePost(cross, "slug1", "title1", "body1", "2022-05-20T13:17:22Z", "lou", "tag1 tag2")
	// 	UpdatePostBySlug(cur, "slug2", newTitle, "body2++", newDate, "lou", "tag1 tag4")
	// 	got := Render("posts/slug2")
	// 	expected := ``
	// 	assertMDEquals(t, got, expected)

	// 	home := Render("")

	// 	if strings.Count(home, newTitle) != 1 {
	// 		t.Errorf("post not edited properly")
	// 	}
	// }

	// {
	// 	// fix
	// 	println("Deleting a post...")
	// 	title := "example title"
	// 	slug := "testSlug1"
	// 	CreatePost(cross, slug, title, "body1", "2022-05-25T13:17:22Z", "lou", "tag1,tag2")

	// 	got := Render("")

	// 	if !strings.Contains(got, title) {
	// 		t.Errorf("post was not added properly")
	// 	}

	// 	postRender := Render("p/" + slug)

	// 	if !strings.Contains(postRender, title) {
	// 		t.Errorf("post not rendered properly")
	// 	}

	// 	DeletePost(cur, slug)
	// 	got = Render("")

	// 	if strings.Contains(got, title) {
	// 		t.Errorf("post was not removed")
	// 	}

	// 	postRender = Render("p/" + slug)

	// 	assertMDEquals(t, postRender, "404")
	// }
	//
	//	// TODO: ?mode=...
	//
}

func assertMDEquals(t *testing.T, got, expected string) {
	t.Helper()
	expected = strings.TrimSpace(expected)
	got = strings.TrimSpace(got)
	if expected != got {
		t.Errorf("invalid render output.\nexpected %q.\ngot      %q.", expected, got)
	} else {
		println("PASSED")
	}
}
