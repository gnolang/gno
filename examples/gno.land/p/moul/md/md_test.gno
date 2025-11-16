package md_test

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
		{"Bold", func() string { return md.Bold("foo") }, "**foo**"},
		{"Italic", func() string { return md.Italic("foo") }, "*foo*"},
		{"Strikethrough", func() string { return md.Strikethrough("foo") }, "~~foo~~"},
		{"H1", func() string { return md.H1("foo") }, "# foo\n"},
		{"HorizontalRule", md.HorizontalRule, "---\n"},
		{"InlineCode", func() string { return md.InlineCode("foo") }, "`foo`"},
		{"CodeBlock", func() string { return md.CodeBlock("foo") }, "```\nfoo\n```"},
		{"LanguageCodeBlock", func() string { return md.LanguageCodeBlock("go", "foo") }, "```go\nfoo\n```"},
		{"Link", func() string { return md.Link("foo", "http://example.com") }, "[foo](http://example.com)"},
		{"UserLink", func() string { return md.UserLink("moul") }, "[@moul](/u/moul)"},
		{"Image", func() string { return md.Image("foo", "http://example.com") }, "![foo](http://example.com)"},
		{"InlineImageWithLink", func() string { return md.InlineImageWithLink("alt", "image-url", "link-url") }, "[![alt](image-url)](link-url)"},
		{"Footnote", func() string { return md.Footnote("foo", "bar") }, "[foo]: bar"},
		{"Paragraph", func() string { return md.Paragraph("foo") }, "foo\n\n"},
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

func TestUserLink(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"username", "moul", "[@moul](/u/moul)"},
		{"address", "g1blah", "[g1blah](/u/g1blah)"},
		{"username with special chars", "user_name", "[@user\\_name](/u/user_name)"},
		{"address with numbers", "g1abc123", "[g1abc123](/u/g1abc123)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := md.UserLink(tt.input)
			if result != tt.expected {
				t.Errorf("UserLink(%q) = %q, want %q", tt.input, result, tt.expected)
			}
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

func TestColumns(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		padded   bool
		expected string
	}{
		{
			name:     "no columns",
			input:    []string{},
			padded:   false,
			expected: "",
		},
		{
			name:     "no columns padded",
			input:    []string{},
			padded:   true,
			expected: "",
		},
		{
			name:   "one column",
			input:  []string{"Column 1"},
			padded: false,
			expected: `<gno-columns>
Column 1
</gno-columns>
`,
		},
		{
			name:   "one column padded",
			input:  []string{"Column 1"},
			padded: true,
			expected: `<gno-columns>
Column 1
|||

|||

|||

</gno-columns>
`,
		},
		{
			name:   "two columns",
			input:  []string{"Column 1", "Column 2"},
			padded: false,
			expected: `<gno-columns>
Column 1
|||
Column 2
</gno-columns>
`,
		},
		{
			name:   "two columns padded",
			input:  []string{"Column 1", "Column 2"},
			padded: true,
			expected: `<gno-columns>
Column 1
|||
Column 2
|||

|||

</gno-columns>
`,
		},
		{
			name:   "four columns",
			input:  []string{"A", "B", "C", "D"},
			padded: false,
			expected: `<gno-columns>
A
|||
B
|||
C
|||
D
</gno-columns>
`,
		},
		{
			name:   "four columns padded",
			input:  []string{"A", "B", "C", "D"},
			padded: true,
			expected: `<gno-columns>
A
|||
B
|||
C
|||
D
</gno-columns>
`,
		},
		{
			name:   "more than four columns",
			input:  []string{"1", "2", "3", "4", "5"},
			padded: false,
			expected: `<gno-columns>
1
|||
2
|||
3
|||
4
|||
5
</gno-columns>
`,
		},
		{
			name:   "more than four columns padded",
			input:  []string{"1", "2", "3", "4", "5"},
			padded: true,
			expected: `<gno-columns>
1
|||
2
|||
3
|||
4
|||
5
|||

|||

|||

</gno-columns>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := md.Columns(tt.input, tt.padded)
			if result != tt.expected {
				t.Errorf("Columns(%v, %v) =\n%q\nwant:\n%q", tt.input, tt.padded, result, tt.expected)
			}
		})
	}
}

func TestColumnsN(t *testing.T) {
	tests := []struct {
		name       string
		content    []string
		colsPerRow int
		padded     bool
		expected   string
	}{
		{
			name:       "empty input",
			content:    []string{},
			colsPerRow: 2,
			padded:     false,
			expected:   "",
		},
		{
			name:       "colsPerRow <= 0",
			content:    []string{"A", "B", "C"},
			colsPerRow: 0,
			padded:     false,
			expected: `<gno-columns>
A
|||
B
|||
C
</gno-columns>
`,
		},
		{
			name:       "exact full row, no padding",
			content:    []string{"A", "B"},
			colsPerRow: 2,
			padded:     false,
			expected: `<gno-columns>
A
|||
B
</gno-columns>
`,
		},
		{
			name:       "partial last row, no padding",
			content:    []string{"A", "B", "C"},
			colsPerRow: 2,
			padded:     false,
			expected: `<gno-columns>
A
|||
B
</gno-columns>
<gno-columns>
C
</gno-columns>
`,
		},
		{
			name:       "partial last row, with padding",
			content:    []string{"A", "B", "C"},
			colsPerRow: 2,
			padded:     true,
			expected: `<gno-columns>
A
|||
B
</gno-columns>
<gno-columns>
C
|||

</gno-columns>
`,
		},
		{
			name:       "padded with more empty cells",
			content:    []string{"X"},
			colsPerRow: 3,
			padded:     true,
			expected: `<gno-columns>
X
|||

|||

</gno-columns>
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := md.ColumnsN(tt.content, tt.colsPerRow, tt.padded)
			if result != tt.expected {
				t.Errorf("ColumnsN(%v, %d, %v) =\n%q\nwant:\n%q", tt.content, tt.colsPerRow, tt.padded, result, tt.expected)
			}
		})
	}
}
