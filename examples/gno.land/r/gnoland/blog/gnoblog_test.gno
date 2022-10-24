package gnoblog

import "testing"

func TestPackage(t *testing.T) {
	// by default, no posts.
	{
		got := Render("/")
		expected := `# Gnoland's Blog

No posts.`
		assertEquals(t, got, expected)
	}

	{
		AdminNewPost("slug1", "title1", "body1", "tag1,tag2")
		AdminNewPost("slug2", "title2", "body2", "tag1,tag3")
		got := Render("/")
		expected := `# Gnoland's Blog

## [title1](r/gnoland/blog:p/slug1)
body1
[tag1](r/gnoland/blog:t/tag1) [tag2](r/gnoland/blog:t/tag2)
## [title2](r/gnoland/blog:p/slug2)
body2
[tag1](r/gnoland/blog:t/tag1) [tag3](r/gnoland/blog:t/tag3)
`
		assertEquals(t, got, expected)
	}
}

func assertEquals(t *testing.T, got, expected string) {
	t.Helper()
	if expected != got {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}
