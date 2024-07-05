package diff

import (
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
			name:     "non-ascii",
			old:      "ASCII 문자가 아닌 것도 되나?",
			new:      "ASCII 문자가 아닌 것도 됨.",
			expected: "ASCII 문자가 아닌 것도 [-되나?][+됨.]",
		},
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
