package md_test

import (
	"testing"

	"gno.land/p/moul/md"
)

func TestHelpers(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{"Bold", func() string { return md.Bold("foo") }, "**foo**"},
		{"Italic", func() string { return md.Italic("foo") }, "*foo*"},
		{"Strikethrough", func() string { return md.Strikethrough("foo") }, "~~foo~~"},
		{"H1", func() string { return md.H1("foo") }, "# foo\n"},
		{"HorizontalRule", md.HorizontalRule, "---\n"},
		{"InlineCode", func() string { return md.InlineCode("foo") }, "`foo`"},
		{"CodeBlock", func() string { return md.CodeBlock("foo") }, "```\nfoo\n```"},
		{"LanguageCodeBlock", func() string { return md.LanguageCodeBlock("go", "foo") }, "```go\nfoo\n```"},
		{"Link", func() string { return md.Link("foo", "http://example.com") }, "[foo](http://example.com)"},
		{"Image", func() string { return md.Image("foo", "http://example.com") }, "![foo](http://example.com)"},
		{"InlineImageWithLink", func() string { return md.InlineImageWithLink("alt", "image-url", "link-url") }, "[![alt](image-url)](link-url)"},
		{"Footnote", func() string { return md.Footnote("foo", "bar") }, "[foo]: bar"},
		{"Paragraph", func() string { return md.Paragraph("foo") }, "foo\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()
			if result != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestLists(t *testing.T) {
	t.Run("BulletList", func(t *testing.T) {
		items := []string{"foo", "bar"}
		expected := "- foo\n- bar\n"
		result := md.BulletList(items)
		if result != expected {
			t.Errorf("BulletList(%q) = %q, want %q", items, result, expected)
		}
	})

	t.Run("OrderedList", func(t *testing.T) {
		items := []string{"foo", "bar"}
		expected := "1. foo\n2. bar\n"
		result := md.OrderedList(items)
		if result != expected {
			t.Errorf("OrderedList(%q) = %q, want %q", items, result, expected)
		}
	})

	t.Run("TodoList", func(t *testing.T) {
		items := []string{"foo", "bar\nmore bar"}
		done := []bool{true, false}
		expected := "- [x] foo\n- [ ] bar\n  more bar\n"
		result := md.TodoList(items, done)
		if result != expected {
			t.Errorf("TodoList(%q, %q) = %q, want %q", items, done, result, expected)
		}
	})
}

func TestNested(t *testing.T) {
	t.Run("Nested Single Level", func(t *testing.T) {
		content := "- foo\n- bar"
		expected := "  - foo\n  - bar"
		result := md.Nested(content, "  ")
		if result != expected {
			t.Errorf("Nested(%q) = %q, want %q", content, result, expected)
		}
	})

	t.Run("Nested Double Level", func(t *testing.T) {
		content := "  - foo\n  - bar"
		expected := "    - foo\n    - bar"
		result := md.Nested(content, "  ")
		if result != expected {
			t.Errorf("Nested(%q) = %q, want %q", content, result, expected)
		}
	})
}
