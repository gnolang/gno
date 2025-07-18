package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnescapeMarkdown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no escaped characters",
			input:    "This is normal text",
			expected: "This is normal text",
		},
		{
			name:     "escaped asterisk",
			input:    "This is \\*bold\\* text",
			expected: "This is *bold* text",
		},
		{
			name:     "escaped underscore",
			input:    "This is \\_italic\\_ text",
			expected: "This is _italic_ text",
		},
		{
			name:     "escaped brackets",
			input:    "This is \\[link\\] text",
			expected: "This is [link] text",
		},
		{
			name:     "escaped parentheses",
			input:    "This is \\(parentheses\\) text",
			expected: "This is (parentheses) text",
		},
		{
			name:     "escaped tilde",
			input:    "This is \\~strikethrough\\~ text",
			expected: "This is ~strikethrough~ text",
		},
		{
			name:     "escaped greater than",
			input:    "This is \\>quote\\> text",
			expected: "This is >quote> text",
		},
		{
			name:     "escaped pipe",
			input:    "This is \\|table\\| text",
			expected: "This is |table| text",
		},
		{
			name:     "escaped minus",
			input:    "This is \\-list\\- text",
			expected: "This is -list- text",
		},
		{
			name:     "escaped plus",
			input:    "This is \\+plus\\+ text",
			expected: "This is +plus+ text",
		},
		{
			name:     "escaped dot",
			input:    "This is \\.dot\\. text",
			expected: "This is .dot. text",
		},
		{
			name:     "escaped exclamation",
			input:    "This is \\!important\\! text",
			expected: "This is !important! text",
		},
		{
			name:     "escaped backtick",
			input:    "This is \\`code\\` text",
			expected: "This is `code` text",
		},
		{
			name:     "multiple escaped characters",
			input:    "\\*bold\\* and \\_italic\\_ and \\`code\\`",
			expected: "*bold* and _italic_ and `code`",
		},
		{
			name:     "mixed escaped and normal characters",
			input:    "Normal text with \\*bold\\* and normal again",
			expected: "Normal text with *bold* and normal again",
		},
		{
			name:     "consecutive escaped characters",
			input:    "\\*\\_\\`\\[\\]\\(",
			expected: "*_`[](",
		},
		{
			name:     "escaped characters at boundaries",
			input:    "\\*start\\* and \\_end\\_",
			expected: "*start* and _end_",
		},
		{
			name:     "real world example from issue #4417",
			input:    "Special char is \\`\\_\\` and \\*bold\\*",
			expected: "Special char is `_` and *bold*",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := UnescapeMarkdown(tc.input)
			assert.Equal(t, tc.expected, result, "UnescapeMarkdown(%q) = %q, want %q", tc.input, result, tc.expected)
		})
	}
}

func TestUnescapeMarkdown_EdgeCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single escaped character",
			input:    "\\*",
			expected: "*",
		},
		{
			name:     "only escaped characters",
			input:    "\\*\\_\\`",
			expected: "*_`",
		},
		{
			name:     "backslash not followed by special character",
			input:    "This is \\normal\\ text",
			expected: "This is \\normal\\ text",
		},
		{
			name:     "multiple backslashes",
			input:    "This is \\\\*not escaped\\*",
			expected: "This is \\*not escaped*",
		},
		{
			name:     "backslash at end",
			input:    "Text with backslash\\",
			expected: "Text with backslash\\",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := UnescapeMarkdown(tc.input)
			assert.Equal(t, tc.expected, result, "UnescapeMarkdown(%q) = %q, want %q", tc.input, result, tc.expected)
		})
	}
}
