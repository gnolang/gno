package gnomod

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSource_Validate(t *testing.T) {
	const gh = "https://github.com/gnolang/gno"
	cases := []struct {
		name   string
		source Source
		errMsg string
	}{
		{"empty", Source{}, ""},
		{"repository only", Source{Repository: "https://github.com/gnolang/gno"}, ""},
		{"full", Source{Repository: "https://github.com/gnolang/gno", Path: "examples/gno.land/p/nt/avl", Revision: "3cc494ec4"}, ""},
		{"any https host", Source{Repository: "https://git.example.org/team/repo"}, ""},
		{"port", Source{Repository: "https://git.example.org:8443/team/repo"}, ""},
		{"nested gitlab groups", Source{Repository: "https://gitlab.com/group/sub/repo.git"}, ""},
		{"sha256 revision", Source{Repository: "https://github.com/gnolang/gno", Revision: strings.Repeat("a", 64)}, ""},
		{"path without repository", Source{Path: "r/demo"}, "'source.repository' is required"},
		{"http", Source{Repository: "http://github.com/gnolang/gno"}, "must be an https URL"},
		{"uppercase scheme", Source{Repository: "HTTPS://github.com/gnolang/gno"}, "must be an https URL"},
		{"javascript", Source{Repository: "javascript:alert(1)"}, "must be an https URL"},
		{"ssh remote", Source{Repository: "git@github.com:gnolang/gno.git"}, "must be an https URL"},
		{"no host", Source{Repository: "https:///gnolang/gno"}, "lowercase DNS host"},
		{"uppercase host", Source{Repository: "https://GitHub.com/gnolang/gno"}, "lowercase DNS host"},
		{"dotless host", Source{Repository: "https://github-work/gnolang/gno"}, "lowercase DNS host"},
		{"drive letter host", Source{Repository: `https://C/\repo`}, "lowercase DNS host"},
		{"homoglyph host", Source{Repository: "https://g\u0456thub.com/gnolang/gno"}, "lowercase DNS host"},
		{"credentials", Source{Repository: "https://x-access-token:secret@github.com/gnolang/gno"}, "lowercase DNS host"},
		{"user only", Source{Repository: "https://token@github.com/gnolang/gno"}, "lowercase DNS host"},
		{"bad port", Source{Repository: "https://github.com:0/gnolang/gno"}, "invalid port"},
		{"double colon port", Source{Repository: "https://github.com:1:2/gnolang/gno"}, "invalid port"},
		{"host only", Source{Repository: "https://github.com"}, "must have a path"},
		{"trailing slash", Source{Repository: "https://github.com/gnolang/gno/"}, "must have a path"},
		{"query", Source{Repository: "https://github.com/gnolang/gno?tab=readme"}, "must have a path"},
		{"fragment", Source{Repository: "https://github.com/gnolang/gno#readme"}, "must have a path"},
		{"percent", Source{Repository: "https://github.com/gnolang/%2e%2e"}, "must have a path"},
		{"dotdot repository", Source{Repository: "https://github.com/gnolang/../evil/repo"}, "must have a path"},
		{"control char", Source{Repository: "https://github.com/gnolang/gno\n"}, "must have a path"},
		{"bidi override", Source{Repository: "https://github.com/gnolang/gno\u202e"}, "must have a path"},
		{"zero width", Source{Repository: "https://github.com/gnolang/g\u200bno"}, "must have a path"},
		{"too long", Source{Repository: "https://github.com/" + strings.Repeat("a", 300)}, "exceeds 256 bytes"},
		{"absolute path", Source{Repository: gh, Path: "/examples"}, "clean relative"},
		{"dotdot path", Source{Repository: gh, Path: "../x"}, "clean relative"},
		{"inner dotdot path", Source{Repository: gh, Path: "a/../../x"}, "clean relative"},
		{"dot path", Source{Repository: gh, Path: "./x"}, "clean relative"},
		{"percent dotdot path", Source{Repository: gh, Path: "%2e%2e/%2e%2e/evil/repo"}, "clean relative"},
		{"query path", Source{Repository: gh, Path: "x?y"}, "clean relative"},
		{"fragment path", Source{Repository: gh, Path: "x#y"}, "clean relative"},
		{"unclean path", Source{Repository: gh, Path: "examples//p"}, "clean relative"},
		{"trailing slash path", Source{Repository: gh, Path: "examples/"}, "clean relative"},
		{"backslash path", Source{Repository: gh, Path: `examples\p`}, "clean relative"},
		{"short revision", Source{Repository: gh, Revision: "abc"}, "lowercase hex commit id"},
		{"branch revision", Source{Repository: gh, Revision: "master"}, "lowercase hex commit id"},
		{"uppercase revision", Source{Repository: gh, Revision: "3CC494EC4"}, "lowercase hex commit id"},
		{"too long revision", Source{Repository: gh, Revision: strings.Repeat("a", 65)}, "lowercase hex commit id"},
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
		pinned bool
	}{
		{"empty", Source{}, "", false},
		{"invalid", Source{Repository: "javascript:alert(1)", Path: "x"}, "", false},
		{"repository only", Source{Repository: "https://github.com/gnolang/gno"}, "https://github.com/gnolang/gno", false},
		{"github path", Source{Repository: "https://github.com/gnolang/gno", Path: "examples/gno.land/p/nt/avl"}, "https://github.com/gnolang/gno/tree/HEAD/examples/gno.land/p/nt/avl", false},
		{"github revision", Source{Repository: "https://github.com/gnolang/gno", Path: "examples", Revision: "3cc494ec4"}, "https://github.com/gnolang/gno/tree/3cc494ec4/examples", true},
		{"github revision only", Source{Repository: "https://github.com/gnolang/gno", Revision: "3cc494ec4"}, "https://github.com/gnolang/gno/tree/3cc494ec4", true},
		{"github .git", Source{Repository: "https://github.com/gnolang/gno.git", Path: "examples"}, "https://github.com/gnolang/gno/tree/HEAD/examples", false},
		{"github port", Source{Repository: "https://github.com:443/gnolang/gno", Path: "examples"}, "https://github.com:443/gnolang/gno/tree/HEAD/examples", false},
		{"github owner only", Source{Repository: "https://github.com/gnolang", Path: "examples"}, "https://github.com/gnolang", false},
		{"gitlab", Source{Repository: "https://gitlab.com/team/repo", Path: "r/x", Revision: "abcdef0"}, "https://gitlab.com/team/repo/-/tree/abcdef0/r/x", true},
		{"gitlab nested", Source{Repository: "https://gitlab.com/group/sub/repo.git", Path: "r/x"}, "https://gitlab.com/group/sub/repo/-/tree/HEAD/r/x", false},
		{"unknown forge", Source{Repository: "https://git.example.org/team/repo", Path: "r/x", Revision: "abcdef0"}, "https://git.example.org/team/repo", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, pinned := tc.source.URL()
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.pinned, pinned)
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

// An invalid section must not make the whole file unreadable: read paths
// (gnodev, gnoweb, the keeper re-reading stored packages) ignore it, and only
// write paths call Source.Validate.
func TestSource_InvalidIgnoredByParse(t *testing.T) {
	f, err := ParseBytes("gnomod.toml", []byte("module = \"gno.land/r/demo/x\"\n[source]\nrepository = \"git@github.com:someone/contracts.git\"\n"))
	require.NoError(t, err)
	assert.Error(t, f.Source.Validate())
	link, _ := f.Source.URL()
	assert.Empty(t, link)
}
