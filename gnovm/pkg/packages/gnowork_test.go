package packages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGnoworkPaths(t *testing.T) {
	tcs := []struct {
		name    string
		gnowork string
		expect  map[string]string
	}{
		{
			name: "common",
			gnowork: `
				paths = [
					["p", "gno.land/p/testns"],
					["r", "gno.land/r/testns"],
				]
			`,
			expect: map[string]string{
				"gno.land/p/testns/somepkg":          "p/somepkg",
				"gno.land/r/testns/somepkg":          "r/somepkg",
				"gno.land/p/testns/somepkg/sub/deep": "p/somepkg/sub/deep",
				"gno.land/p/testns":                  "p",

				"gno.land/p/testnswithsuffix": "gno.land/p/testnswithsuffix", // important case, should not partially match
				"gno.land/p":                  "gno.land/p",
				"gno.other/p/testns/somepkg":  "gno.other/p/testns/somepkg",
			},
		},
		{
			name:    "empty",
			gnowork: "",
			expect: map[string]string{
				"gno.land/p/testnsother":     "gno.land/p/testnsother",
				"gno.land/p":                 "gno.land/p",
				"gno.other/p/testns/somepkg": "gno.other/p/testns/somepkg",
			},
		},
		{
			name: "root",
			gnowork: `
				paths = [
					["", "gno.land"],
				]
			`,
			expect: map[string]string{
				"gno.land/p/testnsother":     "p/testnsother",
				"gno.land/p":                 "p",
				"gno.other/p/testns/somepkg": "gno.other/p/testns/somepkg",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			gw, err := ParseGnowork(tc.name, []byte(tc.gnowork))
			require.NoError(t, err)
			for pkgPath, expected := range tc.expect {
				res := gw.PkgLocalPath(pkgPath)
				require.Equal(t, expected, res)
			}
		})
	}
}
