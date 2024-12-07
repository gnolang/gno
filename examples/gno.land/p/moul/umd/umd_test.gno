package umd

import (
	"testing"

	"gno.land/p/moul/umd"
)

func TestHelpers(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{"Bold", func() string { return umd.Bold("foo") }, "**foo**"},
		{"Italic", func() string { return umd.Italic("foo") }, "*foo*"},
		{"Strikethrough", func() string { return umd.Strikethrough("foo") }, "~~foo~~"},
		{"H1", func() string { return umd.H1("foo") }, "# foo\n"},
		{"HorizontalRule", umd.HorizontalRule, "---\n"},
		{"InlineCode", func() string { return umd.InlineCode("foo") }, "`foo`"},
		{"CodeBlock", func() string { return umd.CodeBlock("foo") }, "```\nfoo\n```"},
		{"LanguageCodeBlock", func() string { return umd.LanguageCodeBlock("go", "foo") }, "```go\nfoo\n```"},
		{"Link", func() string { return umd.Link("foo", "http://example.com") }, "[foo](http://example.com)"},
		{"Image", func() string { return umd.Image("foo", "http://example.com") }, "![foo](http://example.com)"},
		{"InlineImageWithLink", func() string { return umd.InlineImageWithLink("alt", "image-url", "link-url") }, "[![alt](image-url)](link-url)"},
		{"Footnote", func() string { return umd.Footnote("foo", "bar") }, "[foo]: bar"},
		{"Paragraph", func() string { return umd.Paragraph("foo") }, "foo\n"},
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
		result := umd.BulletList(items)
		if result != expected {
			t.Errorf("BulletList(%q) = %q, want %q", items, result, expected)
		}
	})

	t.Run("OrderedList", func(t *testing.T) {
		items := []string{"foo", "bar"}
		expected := "1. foo\n2. bar\n"
		result := umd.OrderedList(items)
		if result != expected {
			t.Errorf("OrderedList(%q) = %q, want %q", items, result, expected)
		}
	})

	t.Run("TodoList", func(t *testing.T) {
		items := []string{"foo", "bar\nmore bar"}
		done := []bool{true, false}
		expected := "- [x] foo\n- [ ] bar\n  more bar\n"
		result := umd.TodoList(items, done)
		if result != expected {
			t.Errorf("TodoList(%q, %q) = %q, want %q", items, done, result, expected)
		}
	})
}

func TestNested(t *testing.T) {
	t.Run("Nested Single Level", func(t *testing.T) {
		content := "- foo\n- bar"
		expected := "  - foo\n  - bar"
		result := umd.Nested(content, "  ")
		if result != expected {
			t.Errorf("Nested(%q) = %q, want %q", content, result, expected)
		}
	})

	t.Run("Nested Double Level", func(t *testing.T) {
		content := "  - foo\n  - bar"
		expected := "    - foo\n    - bar"
		result := umd.Nested(content, "  ")
		if result != expected {
			t.Errorf("Nested(%q) = %q, want %q", content, result, expected)
		}
	})
}
