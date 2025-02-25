// Inspired by: https://cs.opensource.google/go/x/tools/+/master:gopls/internal/test/integration/fake/glob/glob_test.go

package packages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern, input string
		want           bool
	}{
		// Basic cases.
		{"", "", true},
		{"", "a", false},
		{"", "/", false},
		{"abc", "abc", true},

		// ** behavior
		{"**", "abc", true},
		{"**/abc", "abc", true},
		{"**", "abc/def", true},

		// * behavior
		{"/*", "/a", true},
		{"*", "foo", true},
		{"*o", "foo", true},
		{"*o", "foox", false},
		{"f*o", "foo", true},
		{"f*o", "fo", true},

		// Dirs cases
		{"**/", "path/to/foo/", true},
		{"**/", "path/to/foo", true},

		{"path/to/foo", "path/to/foo", true},
		{"path/to/foo", "path/to/bar", false},
		{"path/*/foo", "path/to/foo", true},
		{"path/*/1/*/3/*/5*/foo", "path/to/1/2/3/4/522/foo", true},
		{"path/*/1/*/3/*/5*/foo", "path/to/1/2/3/4/722/foo", false},
		{"path/*/1/*/3/*/5*/foo", "path/to/1/2/3/4/522/bar", false},
		{"path/*/foo", "path/to/to/foo", false},
		{"path/**/foo", "path/to/to/foo", true},
		{"path/**/foo", "path/to/to/bar", false},
		{"path/**/foo", "path/foo", true},
		{"**/abc/**", "foo/r/x/abc/bar", true},

		// Realistic examples.
		{"**/*.ts", "path/to/foo.ts", true},
		{"**/*.js", "path/to/foo.js", true},
		{"**/*.go", "path/to/foo.go", true},
	}

	for _, test := range tests {
		g, err := Parse(test.pattern)
		require.NoErrorf(t, err, "Parse(%q) failed unexpectedly: %v", test.pattern, err)
		assert.Equalf(t, test.want, g.Match(test.input),
			"Parse(%q).Match(%q) = %t, want %t", test.pattern, test.input, !test.want, test.want)
	}
}

func TestBaseFreeStar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern, baseFree string
	}{
		// Basic cases.
		{"", ""},
		{"foo", "foo"},
		{"foo/bar", "foo/bar"},
		{"foo///bar", "foo/bar"},
		{"foo/bar/", "foo/bar/"},
		{"foo/bar/*/*/z", "foo/bar/"},
		{"foo/bar/**", "foo/bar/"},
		{"**", ""},
		{"/**", "/"},
	}

	for _, test := range tests {
		g, err := Parse(test.pattern)
		require.NoErrorf(t, err, "Parse(%q) failed unexpectedly: %v", test.pattern, err)
		got := g.StarFreeBase()
		assert.Equalf(t, test.baseFree, got,
			"Parse(%q).Match(%q) = %q, want %q", test.pattern, test.baseFree, got, test.baseFree)
	}
}
