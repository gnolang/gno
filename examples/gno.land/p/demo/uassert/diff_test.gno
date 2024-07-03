package uassert

import (
	"testing"
)

func TestMyersDiff(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			diff := myersDiff(tc.old, tc.new)
			result := formatDiff(diff)
			if result != tc.expected {
				t.Errorf("Expected: %s, got: %s", tc.expected, result)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name        string
		expected    interface{}
		actual      interface{}
		shouldPass  bool
		expectedMsg string
	}{
		{
			name:        "Equal strings",
			expected:    "abc",
			actual:      "abc",
			shouldPass:  true,
			expectedMsg: "",
		},
		{
			name:        "Different strings",
			expected:    "abc",
			actual:      "abd",
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: strings are different\n\tDiff: ab[-c][+d]",
		},
		{
			name:        "Equal integers",
			expected:    123,
			actual:      123,
			shouldPass:  true,
			expectedMsg: "",
		},
		{
			name:        "Different integers",
			expected:    123,
			actual:      456,
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: same type but different value\n\texpected: 123\n\tactual:   456",
		},
		{
			name:        "Different types",
			expected:    "123",
			actual:      123,
			shouldPass:  false,
			expectedMsg: "error: uassert.Equal: different types",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockT := &mockTestingT{}
			result := Equal(mockT, tc.expected, tc.actual)
			if result != tc.shouldPass {
				t.Errorf("Expected Equal to return %v, but got %v", tc.shouldPass, result)
			}
			if tc.shouldPass {
				mockT.empty(t)
			} else {
				mockT.equals(t, tc.expectedMsg)
			}
		})
	}
}
