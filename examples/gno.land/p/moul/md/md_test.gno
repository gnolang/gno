package md

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestHelpers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		// Basic formatting
		{"Bold", func() string { return Bold("foo") }, "**foo**"},
		{"Bold with spaces", func() string { return Bold("  foo  ") }, "**foo**"},
		{"Italic", func() string { return Italic("foo") }, "*foo*"},
		{"Italic with spaces", func() string { return Italic("  foo  ") }, "*foo*"},
		{"Strikethrough", func() string { return Strikethrough("foo") }, "~~foo~~"},

		// Headers
		{"H1", func() string { return H1("foo") }, "# foo"},
		{"H2", func() string { return H2("foo") }, "## foo"},
		{"H3", func() string { return H3("foo") }, "### foo"},
		{"H4", func() string { return H4("foo") }, "#### foo"},
		{"H5", func() string { return H5("foo") }, "##### foo"},
		{"H6", func() string { return H6("foo") }, "###### foo"},

		// Headers with spaces
		{"H1 with spaces", func() string { return H1("  foo  ") }, "# foo"},
		{"H2 with spaces", func() string { return H2("  foo  ") }, "## foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			uassert.Equal(t, tt.expected, got)
		})
	}
}

func TestLists(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{"BulletItem", func() string { return BulletItem("foo") }, "- foo\n"},
		{"BulletItem with spaces", func() string { return BulletItem("  foo  ") }, "- foo\n"},
		{"BulletItem multiline", func() string { return BulletItem("  foo  \n  bar  ") }, "- foo\n  bar\n"},
		{"BulletList", func() string { return BulletList([]string{"foo", "bar"}) }, "- foo\n- bar\n"},
		{"BulletList with spaces", func() string { return BulletList([]string{"  foo  ", "  bar  "}) }, "- foo\n- bar\n"},
		{"OrderedList", func() string { return OrderedList([]string{"foo", "bar"}) }, "1. foo\n2. bar\n"},
		{"OrderedList with spaces", func() string { return OrderedList([]string{"  foo  ", "  bar  "}) }, "1. foo\n2. bar\n"},
		{"TodoList", func() string { return TodoList([]string{"foo", "bar"}, []bool{true, false}) }, "- [x] foo\n- [ ] bar\n"},
		{"TodoList with spaces", func() string { return TodoList([]string{"  foo  ", "  bar  \n  more  "}, []bool{true, false}) }, "- [x] foo\n- [ ] bar\n  more\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			uassert.Equal(t, tt.expected, got)
		})
	}
}

func TestFormatting(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{"Blockquote", func() string { return Blockquote("foo") }, "> foo"},
		{"Blockquote with spaces", func() string { return Blockquote("  foo  \n  bar  ") }, "> foo\n> bar"},
		{"Link", func() string { return Link("foo", "http://example.com") }, "[foo](http://example.com)"},
		{"Link with spaces", func() string { return Link("  foo  ", "http://example.com") }, "[foo](http://example.com)"},
		{"Image", func() string { return Image("foo", "http://example.com") }, "![foo](http://example.com)"},
		{"Image with spaces", func() string { return Image("  foo  ", "http://example.com") }, "![foo](http://example.com)"},
		{"Paragraph", func() string { return Paragraph("foo") }, "foo"},
		{"Paragraph with spaces", func() string { return Paragraph("  foo  ") }, "foo"},
		{"Paragraph multiline", func() string { return Paragraph("  foo  \n  bar  ") }, "foo\n  bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			uassert.Equal(t, tt.expected, got)
		})
	}
}

func TestNested(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		prefix   string
		expected string
	}{
		{
			name:     "basic nesting",
			content:  "- foo\n- bar",
			prefix:   "  ",
			expected: "  - foo\n  - bar",
		},
		{
			name:     "preserve spaces",
			content:  "  - foo  \n  - bar  ",
			prefix:   "  ",
			expected: "    - foo  \n    - bar  ",
		},
		{
			name:     "empty lines",
			content:  "- foo\n\n- bar",
			prefix:   "  ",
			expected: "  - foo\n\n  - bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Nested(tt.content, tt.prefix)
			uassert.Equal(t, tt.expected, got)
		})
	}
}
