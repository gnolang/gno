package packages

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// fakeRPCFetcher implements pkgdownload.RPCPackageFetcher and records what it received.
type fakeRPCFetcher struct {
	applied map[string]string
}

func (f *fakeRPCFetcher) FetchPackage(string) ([]*std.MemFile, error) { return nil, nil }
func (f *fakeRPCFetcher) OverrideDomainsRPCs(m map[string]string)     { f.applied = m }

// fakePlainFetcher implements only pkgdownload.PackageFetcher (no override support).
type fakePlainFetcher struct{}

func (fakePlainFetcher) FetchPackage(string) ([]*std.MemFile, error) { return nil, nil }

func TestApplyRPCOverrides(t *testing.T) {
	t.Run("no overrides is a no-op even for a plain fetcher", func(t *testing.T) {
		require.NoError(t, applyRPCOverrides(fakePlainFetcher{}, nil))
	})

	t.Run("overrides are pushed into an rpc-capable fetcher", func(t *testing.T) {
		f := &fakeRPCFetcher{}
		overrides := map[string]string{"gno.land": "http://localhost:26657"}
		require.NoError(t, applyRPCOverrides(f, overrides))
		require.Equal(t, overrides, f.applied)
	})

	t.Run("overrides on an unsupported fetcher error instead of being dropped", func(t *testing.T) {
		err := applyRPCOverrides(fakePlainFetcher{}, map[string]string{"gno.land": "http://localhost:26657"})
		require.ErrorContains(t, err, "does not support")
	})
}
