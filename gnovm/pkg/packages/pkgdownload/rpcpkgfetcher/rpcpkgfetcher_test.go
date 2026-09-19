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

func TestOverrideDomainsRPCs(t *testing.T) {
	t.Run("populates a nil override map", func(t *testing.T) {
		gpf := &gnoPackageFetcher{}
		gpf.OverrideDomainsRPCs(map[string]string{"gno.land": "http://localhost:26657"})

		res, err := rpcURLFromPkgPath("gno.land/p/nt/avl/v0", gpf.remoteOverrides)
		require.NoError(t, err)
		require.Equal(t, "http://localhost:26657", res)
	})

	t.Run("merges onto existing overrides", func(t *testing.T) {
		gpf := &gnoPackageFetcher{remoteOverrides: map[string]string{"gno.land": "https://old:443"}}
		gpf.OverrideDomainsRPCs(map[string]string{
			"gno.land":    "http://localhost:26657",
			"example.com": "http://localhost:8080",
		})
		require.Equal(t, map[string]string{
			"gno.land":    "http://localhost:26657",
			"example.com": "http://localhost:8080",
		}, gpf.remoteOverrides)
	})

	t.Run("empty input is a no-op", func(t *testing.T) {
		gpf := &gnoPackageFetcher{}
		gpf.OverrideDomainsRPCs(nil)
		require.Nil(t, gpf.remoteOverrides)
	})
}
