package umd_test

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
		{
			name:     "Bold",
			function: func() string { return umd.Bold("foo") },
			expected: "**foo**",
		},
		{
			name:     "Italic",
			function: func() string { return umd.Italic("foo") },
			expected: "*foo*",
		},
		{
			name:     "Strikethrough",
			function: func() string { return umd.Strikethrough("foo") },
			expected: "~~foo~~",
		},
		{
			name:     "H1",
			function: func() string { return umd.H1("foo") },
			expected: "# foo\n",
		},
		{
			name:     "HorizontalRule",
			function: umd.HorizontalRule,
			expected: "---\n",
		},
		{
			name: "InlineCode",
			function: func() string {
				return umd.InlineCode("foo")
			},
			expected: "`foo`",
		},
		{
			name: "CodeBlock",
			function: func() string {
				return umd.CodeBlock("foo")
			},
			expected: "```\nfoo\n```",
		},
		{
			name: "LanguageCodeBlock",
			function: func() string {
				return umd.LanguageCodeBlock("go", "foo")
			},
			expected: "```go\nfoo\n```",
		},
		{
			name: "Link",
			function: func() string {
				return umd.Link("foo", "http://example.com")
			},
			expected: "[foo](http://example.com)",
		},
		{
			name: "Image",
			function: func() string {
				return umd.Image("foo", "http://example.com")
			},
			expected: "![foo](http://example.com)",
		},
		{
			name: "Footnote",
			function: func() string {
				return umd.Footnote("foo", "bar")
			},
			expected: "[foo]: bar",
		},
		{
			name: "Paragraph",
			function: func() string {
				return umd.Paragraph("foo")
			},
			expected: "foo\n",
		},
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

func TestBlockquote(t *testing.T) {
	input := "foo\nbar"
	expected := "> foo\n> bar\n"
	result := umd.Blockquote(input)
	if result != expected {
		t.Errorf("Blockquote(%q) = %q, want %q", input, result, expected)
	}
}
