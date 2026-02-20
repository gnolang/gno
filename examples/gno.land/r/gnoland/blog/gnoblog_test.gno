package blog

import (
	"strings"
	"testing"
)

func TestPackage(cur realm, t *testing.T) {
	clearState(t)

	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))

	// by default, no posts.
	{
		got := Render("")
		expected := `
# Gno.land's blog

No posts.
`
		assertMDEquals(t, got, expected)
	}

	// create two posts, list post.
	{
		ModAddPost(cross, "slug1", "title1", "body1", "2022-05-20T13:17:22Z", "moul", "tag1,tag2")
		ModAddPost(cross, "slug2", "title2", "body2", "2022-05-20T13:17:23Z", "moul", "tag1,tag3")
		got := Render("")
		expected := `
			# Gno.land's blog

<gno-columns>
### [title2](/r/gnoland/blog:p/slug2)
20 May 2022

|||

### [title1](/r/gnoland/blog:p/slug1)
20 May 2022

|||
</gno-columns>
`
		assertMDEquals(t, got, expected)
	}

	// view post.
	{
		got := Render("p/slug2")
		expected := `
	<main class='gno-tmpl-page'>

# title2

body2

---

Tags: [#tag1](/r/gnoland/blog:t/tag1) [#tag3](/r/gnoland/blog:t/tag3)

Written by moul on 20 May 2022

Published by g1manfred47kzduec920z88wfr64ylksmdcedlf5 to Gno.land's blog

---
<details><summary>Comment section</summary>

</details>
</main>
	
		`
		assertMDEquals(t, got, expected)
	}

	// list by tags.
	{
		got := Render("t/invalid")
		expected := "# [Gno.land's blog](/r/gnoland/blog:) / t / invalid\n\nNo posts."
		assertMDEquals(t, got, expected)

		got = Render("t/tag2")
		expected = `
# [Gno.land's blog](/r/gnoland/blog:) / t / tag2


### [title1](/r/gnoland/blog:p/slug1)
20 May 2022
		`
		assertMDEquals(t, got, expected)
	}

	// add comments.
	{
		AddComment(cross, "slug1", "comment1")
		AddComment(cross, "slug2", "comment2")
		AddComment(cross, "slug1", "comment3")
		AddComment(cross, "slug2", "comment4")
		AddComment(cross, "slug1", "comment5")
		got := Render("p/slug2")
		expected := `<main class='gno-tmpl-page'>

# title2

body2

---

Tags: [#tag1](/r/gnoland/blog:t/tag1) [#tag3](/r/gnoland/blog:t/tag3)

Written by moul on 20 May 2022

Published by g1manfred47kzduec920z88wfr64ylksmdcedlf5 to Gno.land's blog

---
<details><summary>Comment section</summary>

<h5>comment4

</h5><h6>by g1manfred47kzduec920z88wfr64ylksmdcedlf5 on 13 Feb 09 23:31 UTC</h6>

---

<h5>comment2

</h5><h6>by g1manfred47kzduec920z88wfr64ylksmdcedlf5 on 13 Feb 09 23:31 UTC</h6>

---

</details>
</main>

		`
		assertMDEquals(t, got, expected)
	}

	// edit post.
	{
		oldTitle := "title2"
		oldDate := "2022-05-20T13:17:23Z"

		ModEditPost(cur, "slug2", oldTitle, "body2++", oldDate, "manfred", "tag1,tag4")
		got := Render("p/slug2")
		expected := `<main class='gno-tmpl-page'>

# title2

body2++

---

Tags: [#tag1](/r/gnoland/blog:t/tag1) [#tag4](/r/gnoland/blog:t/tag4)

Written by manfred on 20 May 2022

Published by g1manfred47kzduec920z88wfr64ylksmdcedlf5 to Gno.land's blog

---
<details><summary>Comment section</summary>

<h5>comment4

</h5><h6>by g1manfred47kzduec920z88wfr64ylksmdcedlf5 on 13 Feb 09 23:31 UTC</h6>

---

<h5>comment2

</h5><h6>by g1manfred47kzduec920z88wfr64ylksmdcedlf5 on 13 Feb 09 23:31 UTC</h6>

---

</details>
</main>

		`
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
		ModAddPost(cross, slug, title, "body1", "2022-05-25T13:17:22Z", "moul", "tag1,tag2")

		got := Render("")

		if !strings.Contains(got, title) {
			t.Errorf("post was not added properly")
		}

		postRender := Render("p/" + slug)

		if !strings.Contains(postRender, title) {
			t.Errorf("post not rendered properly")
		}

		ModRemovePost(cur, slug)
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
