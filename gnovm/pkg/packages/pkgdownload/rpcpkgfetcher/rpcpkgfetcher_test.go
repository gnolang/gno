package rpcpkgfetcher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRpcURLFromPkgPath(t *testing.T) {
	cases := []struct {
		name          string
		pkgPath       string
		overrides     map[string]string
		result        string
		errorContains string
	}{
		{
			name:    "happy path simple",
			pkgPath: "gno.land/p/nt/avl/v0",
			result:  "https://rpc.gno.land:443",
		},
		{
			name:      "happy path override",
			pkgPath:   "gno.land/p/nt/avl/v0",
			overrides: map[string]string{"gno.land": "https://example.com/rpc:42"},
			result:    "https://example.com/rpc:42",
		},
		{
			name:      "happy path override no effect",
			pkgPath:   "gno.land/p/nt/avl/v0",
			overrides: map[string]string{"some.chain": "https://example.com/rpc:42"},
			result:    "https://rpc.gno.land:443",
		},
		{
			name:          "error bad pkg path",
			pkgPath:       "std",
			result:        "",
			errorContains: `bad pkg path "std"`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := rpcURLFromPkgPath(c.pkgPath, c.overrides)
			if len(c.errorContains) == 0 {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, c.errorContains)
			}
			require.Equal(t, c.result, res)
		})
	}
}
