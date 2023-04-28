package gnomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModuleDeprecated(t *testing.T) {
	for _, tc := range []struct {
		desc, in, expected string
	}{
		{
			desc: "no_comment",
			in:   `module m`,
		},
		{
			desc: "other_comment",
			in: `// yo
			module m`,
		},
		{
			desc: "deprecated_no_colon",
			in: `//Deprecated
			module m`,
		},
		{
			desc: "deprecated_no_space",
			in: `//Deprecated:blah
			module m`,
			expected: "blah",
		},
		{
			desc: "deprecated_simple",
			in: `// Deprecated: blah
			module m`,
			expected: "blah",
		},
		{
			desc: "deprecated_lowercase",
			in: `// deprecated: blah
			module m`,
		},
		{
			desc: "deprecated_multiline",
			in: `// Deprecated: one
			// two
			module m`,
			expected: "one\ntwo",
		},
		{
			desc: "deprecated_mixed",
			in: `// some other comment
			// Deprecated: blah
			module m`,
		},
		{
			desc: "deprecated_middle",
			in: `// module m is Deprecated: blah
			module m`,
		},
		{
			desc: "deprecated_multiple",
			in: `// Deprecated: a
			// Deprecated: b
			module m`,
			expected: "a\nDeprecated: b",
		},
		{
			desc: "deprecated_paragraph",
			in: `// Deprecated: a
			// b
			//
			// c
			module m`,
			expected: "a\nb",
		},
		{
			desc: "deprecated_paragraph_space",
			in: `// Deprecated: the next line has a space
			// 
			// c
			module m`,
			expected: "the next line has a space",
		},
		{
			desc:     "deprecated_suffix",
			in:       `module m // Deprecated: blah`,
			expected: "blah",
		},
		{
			desc: `deprecated_mixed_suffix`,
			in: `// some other comment
			module m // Deprecated: blah`,
		},
		{
			desc: "deprecated_mixed_suffix_paragraph",
			in: `// some other comment
			//
			module m // Deprecated: blah`,
			expected: "blah",
		},
		{
			desc: "deprecated_block",
			in: `// Deprecated: blah
			module (
				m
			)`,
			expected: "blah",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			f, err := Parse("in", []byte(tc.in))
			assert.Nil(t, err)
			assert.Equal(t, tc.expected, f.Module.Deprecated)
		})
	}
}
