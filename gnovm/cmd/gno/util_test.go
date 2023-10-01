package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		names    []string
		expected []bool
	}{
		{
			pattern:  "foo",
			names:    []string{"foo", "bar", "baz", "foo/bar"},
			expected: []bool{true, false, false, false},
		},
		{
			pattern:  "foo/...",
			names:    []string{"foo", "foo/bar", "foo/bar/baz", "bar", "baz"},
			expected: []bool{true, true, true, false, false},
		},
		{
			pattern:  "foo/bar/...",
			names:    []string{"foo/bar", "foo/bar/baz", "foo/baz/bar", "foo", "bar"},
			expected: []bool{true, true, false, false, false},
		},
		{
			pattern:  "foo/.../baz",
			names:    []string{"foo/bar", "foo/bar/baz", "foo/baz/bar", "foo", "bar"},
			expected: []bool{false, true, false, false, false},
		},
		{
			pattern:  "foo/.../baz/...",
			names:    []string{"foo/bar/baz", "foo/baz/bar", "foo/bar/baz/qux", "foo/baz/bar/qux"},
			expected: []bool{true, false, true, false},
		},
		{
			pattern:  "...",
			names:    []string{"foo", "bar", "baz", "foo/bar", "foo/bar/baz"},
			expected: []bool{true, true, true, true, true},
		},
		{
			pattern:  ".../bar",
			names:    []string{"foo", "bar", "baz", "foo/bar", "foo/bar/baz"},
			expected: []bool{false, false, false, true, false},
		},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			matchFunc := matchPattern(test.pattern)
			for i, name := range test.names {
				res := matchFunc(name)
				assert.Equal(t, test.expected[i], res, "Expected: %v, Got: %v", test.expected[i], res)
			}
		})
	}
}
