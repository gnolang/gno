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
		{"CodeBlock", func() string { return md.CodeBlock("foo") }, "```\nfoo\n```\n"},
		{"LanguageCodeBlock", func() string { return md.LanguageCodeBlock("go", "foo") }, "```go\nfoo\n```\n"},
		{"Link", func() string { return md.Link("foo", "http://example.com") }, "[foo](http://example.com)"},
		{"UserLink", func() string { return md.UserLink("moul") }, "[@moul](/u/moul)"},
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
