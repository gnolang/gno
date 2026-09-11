package diff

import (
	"strings"
	"testing"
)

func TestMyersDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected string
	}{
		{
			name:     "No difference",
			old:      "abc",
			new:      "abc",
			expected: "abc",
		},
		{
			name:     "Simple insertion",
			old:      "ac",
			new:      "abc",
			expected: "a[+b]c",
		},
		{
			name:     "Simple deletion",
			old:      "abc",
			new:      "ac",
			expected: "a[-b]c",
		},
		{
			name:     "Simple substitution",
			old:      "abc",
			new:      "abd",
			expected: "ab[-c][+d]",
		},
		{
			name:     "Multiple changes",
			old:      "The quick brown fox jumps over the lazy dog",
			new:      "The quick brown cat jumps over the lazy dog",
			expected: "The quick brown [-fox][+cat] jumps over the lazy dog",
		},
		{
			name:     "Prefix and suffix",
			old:      "Hello, world!",
			new:      "Hello, beautiful world!",
			expected: "Hello, [+beautiful ]world!",
		},
		{
			name:     "Complete change",
			old:      "abcdef",
			new:      "ghijkl",
			expected: "[-abcdef][+ghijkl]",
		},
		{
			name:     "Empty strings",
			old:      "",
			new:      "",
			expected: "",
		},
		{
			name:     "Old empty",
			old:      "",
			new:      "abc",
			expected: "[+abc]",
		},
		{
			name:     "New empty",
			old:      "abc",
			new:      "",
			expected: "[-abc]",
		},
		{
			name:     "non-ascii (Korean characters)",
			old:      "ASCII 문자가 아닌 것도 되나?",
			new:      "ASCII 문자가 아닌 것도 됨.",
			expected: "ASCII 문자가 아닌 것도 [-되나?][+됨.]",
		},
		{
			name:     "Emoji diff",
			old:      "Hello 👋 World 🌍",
			new:      "Hello 👋 Beautiful 🌸 World 🌍",
			expected: "Hello 👋 [+Beautiful 🌸 ]World 🌍",
		},
		{
			name:     "Mixed multibyte and ASCII",
			old:      "こんにちは World",
			new:      "こんばんは World",
			expected: "こん[-にち][+ばん]は World",
		},
		{
			name:     "Chinese characters",
			old:      "我喜欢编程",
			new:      "我喜欢看书和编程",
			expected: "我喜欢[+看书和]编程",
		},
		{
			name:     "Combining characters",
			old:      "e\u0301", // é (e + ´)
			new:      "e\u0300", // è (e + `)
			expected: "e[-\u0301][+\u0300]",
		},
		{
			name:     "Right-to-Left languages",
			old:      "שלום",
			new:      "שלום עולם",
			expected: "שלום[+ עולם]",
		},
		{
			name:     "Normalization NFC and NFD",
			old:      "e\u0301", // NFD (decomposed)
			new:      "\u00e9",  // NFC (precomposed)
			expected: "[-e\u0301][+\u00e9]",
		},
		{
			name:     "Case sensitivity",
			old:      "abc",
			new:      "Abc",
			expected: "[-a][+A]bc",
		},
		{
			name:     "Surrogate pairs",
			old:      "Hello 🌍",
			new:      "Hello 🌎",
			expected: "Hello [-🌍][+🌎]",
		},
		{
			name:     "Control characters",
			old:      "Line1\nLine2",
			new:      "Line1\r\nLine2",
			expected: "Line1[+\r]\nLine2",
		},
		{
			name:     "Mixed scripts",
			old:      "Hello नमस्ते こんにちは",
			new:      "Hello สวัสดี こんにちは",
			expected: "Hello [-नमस्ते][+สวัสดี] こんにちは",
		},
		{
			name:     "Unicode normalization",
			old:      "é",       // U+00E9 (precomposed)
			new:      "e\u0301", // U+0065 U+0301 (decomposed)
			expected: "[-é][+e\u0301]",
		},
		{
			name:     "Directional marks",
			old:      "Hello\u200Eworld", // LTR mark
			new:      "Hello\u200Fworld", // RTL mark
			expected: "Hello[-\u200E][+\u200F]world",
		},
		{
			name:     "Zero-width characters",
			old:      "ab\u200Bc", // Zero-width space
			new:      "abc",
			expected: "ab[-\u200B]c",
		},
		{
			name:     "Worst-case scenario (completely different strings)",
			old:      strings.Repeat("a", 1000),
			new:      strings.Repeat("b", 1000),
			expected: "[-" + strings.Repeat("a", 1000) + "][+" + strings.Repeat("b", 1000) + "]",
		},
		//{ // disabled for testing performance
		// XXX: consider adding a flag to run such tests, not like `-short`, or switching to a `-bench`, maybe.
		//	name:     "Very long strings",
		//	old:      strings.Repeat("a", 10000) + "b" + strings.Repeat("a", 10000),
		//	new:      strings.Repeat("a", 10000) + "c" + strings.Repeat("a", 10000),
		//	expected: strings.Repeat("a", 10000) + "[-b][+c]" + strings.Repeat("a", 10000),
		//},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diff := MyersDiff(tc.old, tc.new)
			result := Format(diff)
			if result != tc.expected {
				t.Errorf("Expected: %s, got: %s", tc.expected, result)
			}
		})
	}
}
