package gnoblog

import (
	"strings"
	"testing"
)

func TestPackage(t *testing.T) {
	// by default, no posts.
	{
		got := Render("")
		expected := `
# Gnoland's Blog

No posts.
`
		assertMDEquals(t, got, expected)
	}

	// create two posts, list post.
	{
		AdminNewPost("slug1", "title1", "body1", "tag1,tag2")
		AdminNewPost("slug2", "title2", "body2", "tag1,tag3")
		got := Render("")
		expected := `
# Gnoland's Blog

## [title1](r/gnoland/blog:p/slug1)
body1
[tag1](r/gnoland/blog:t/tag1) [tag2](r/gnoland/blog:t/tag2)

## [title2](r/gnoland/blog:p/slug2)
body2
[tag1](r/gnoland/blog:t/tag1) [tag3](r/gnoland/blog:t/tag3)
`
		assertMDEquals(t, got, expected)
	}

	// list posts.
	{
		got := Render("p/slug2")
		expected := `
# Gnoland's Blog

## [title2](r/gnoland/blog:p/slug2)
body2
[tag1](r/gnoland/blog:t/tag1) [tag3](r/gnoland/blog:t/tag3)
`
		assertMDEquals(t, got, expected)
	}

	// TODO: comments.
	// TODO: pagination.
	// TODO: ?format=...

	// TODO: all 404
	{
		notFoundPaths := []string{
			"p/slug3",
			"p",
			"p/",
			"t",
			"t/",
			"/",
			"p/slug1/",
		}
		for _, notFoundPath := range notFoundPaths {
			got := Render(notFoundPath)
			if got != "404" {
				t.Errorf("path %q: expected 404", notFoundPath)
			}
		}
	}
}

func assertMDEquals(t *testing.T, got, expected string) {
	t.Helper()
	expected = strings.TrimSpace(expected)
	got = strings.TrimSpace(got)
	if expected != got {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}
