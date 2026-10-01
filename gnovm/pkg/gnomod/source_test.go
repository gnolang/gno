package gnomod

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSource_Validate(t *testing.T) {
	cases := []struct {
		name   string
		source Source
		errMsg string
	}{
		{"empty", Source{}, ""},
		{"repository only", Source{Repository: "https://github.com/gnolang/gno"}, ""},
		{"full", Source{Repository: "https://github.com/gnolang/gno", Path: "examples/gno.land/p/nt/avl", Revision: "3cc494ec4"}, ""},
		{"any https host", Source{Repository: "https://git.example.org/team/repo"}, ""},
		{"path without repository", Source{Path: "r/demo"}, "'source.repository' is required"},
		{"http", Source{Repository: "http://github.com/gnolang/gno"}, "must be an https URL"},
		{"javascript", Source{Repository: "javascript:alert(1)"}, "must be an https URL"},
		{"ssh remote", Source{Repository: "git@github.com:gnolang/gno.git"}, "'source.repository'"},
		{"no host", Source{Repository: "https:///gnolang/gno"}, "has no host"},
		{"credentials", Source{Repository: "https://x-access-token:secret@github.com/gnolang/gno"}, "must not embed credentials"},
		{"query", Source{Repository: "https://github.com/gnolang/gno?tab=readme"}, "query or fragment"},
		{"fragment", Source{Repository: "https://github.com/gnolang/gno#readme"}, "query or fragment"},
		{"control char", Source{Repository: "https://github.com/gnolang/gno\n"}, "control characters"},
		{"too long", Source{Repository: "https://github.com/" + strings.Repeat("a", 300)}, "exceeds 256 bytes"},
		{"absolute path", Source{Repository: "https://github.com/gnolang/gno", Path: "/examples"}, "clean relative path"},
		{"dotdot path", Source{Repository: "https://github.com/gnolang/gno", Path: "../x"}, "clean relative path"},
		{"unclean path", Source{Repository: "https://github.com/gnolang/gno", Path: "examples//p"}, "clean relative path"},
		{"backslash path", Source{Repository: "https://github.com/gnolang/gno", Path: `examples\p`}, "clean relative path"},
		{"short revision", Source{Repository: "https://github.com/gnolang/gno", Revision: "abc"}, "lowercase hex commit id"},
		{"branch revision", Source{Repository: "https://github.com/gnolang/gno", Revision: "master"}, "lowercase hex commit id"},
		{"uppercase revision", Source{Repository: "https://github.com/gnolang/gno", Revision: "3CC494EC4"}, "lowercase hex commit id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.source.Validate()
			if tc.errMsg == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestSource_URL(t *testing.T) {
	cases := []struct {
		name   string
		source Source
		want   string
	}{
		{"empty", Source{}, ""},
		{"repository only", Source{Repository: "https://github.com/gnolang/gno/"}, "https://github.com/gnolang/gno"},
		{"github path", Source{Repository: "https://github.com/gnolang/gno", Path: "examples/gno.land/p/nt/avl"}, "https://github.com/gnolang/gno/tree/HEAD/examples/gno.land/p/nt/avl"},
		{"github revision", Source{Repository: "https://github.com/gnolang/gno", Path: "examples", Revision: "3cc494ec4"}, "https://github.com/gnolang/gno/tree/3cc494ec4/examples"},
		{"gitlab", Source{Repository: "https://gitlab.com/team/repo", Path: "r/x", Revision: "abcdef0"}, "https://gitlab.com/team/repo/-/tree/abcdef0/r/x"},
		{"unknown forge", Source{Repository: "https://git.example.org/team/repo", Path: "r/x"}, "https://git.example.org/team/repo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.source.URL())
		})
	}
}

func TestSource_RoundTrip(t *testing.T) {
	const in = `module = "gno.land/r/demo/x"
gno = "0.9"

[source]
  repository = "https://github.com/gnolang/gno"
  path = "examples/gno.land/r/demo/x"
  revision = "3cc494ec4"
`
	f, err := ParseBytes("gnomod.toml", []byte(in))
	require.NoError(t, err)
	assert.Equal(t, Source{
		Repository: "https://github.com/gnolang/gno",
		Path:       "examples/gno.land/r/demo/x",
		Revision:   "3cc494ec4",
	}, f.Source)

	// The keeper re-encodes gnomod.toml on every AddPackage (stampGnomod);
	// the section must survive that.
	f2, err := ParseBytes("gnomod.toml", []byte(f.WriteString()))
	require.NoError(t, err)
	assert.Equal(t, f.Source, f2.Source)

	// And a file without one must not grow an empty [source] table.
	f.Source = Source{}
	assert.NotContains(t, f.WriteString(), "source")
}

func TestSource_InvalidRejectedByParse(t *testing.T) {
	_, err := ParseBytes("gnomod.toml", []byte("module = \"gno.land/r/demo/x\"\n[source]\nrepository = \"javascript:alert(1)\"\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be an https URL")
}
