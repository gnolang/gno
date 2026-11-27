package gnovm

import (
	"path/filepath"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/stretchr/testify/require"
)

// TestStdlibsMatchTheTree pins that every stdlib read from the embedded
// sources is byte-identical to the one read from the directory, as genesis
// read it before the embedding: a file type the patterns miss would change
// the genesis state of every new chain.
func TestStdlibsMatchTheTree(t *testing.T) {
	t.Parallel()

	for _, lib := range stdlibs.InitOrder() {
		embedded, err := gno.ReadMemPackageFS(Stdlibs(), lib, lib, gno.MPStdlibAll)
		require.NoError(t, err, lib)
		onDisk, err := gno.ReadMemPackage(filepath.Join("stdlibs", filepath.FromSlash(lib)), lib, gno.MPStdlibAll)
		require.NoError(t, err, lib)
		require.Equal(t, onDisk, embedded, lib)
	}
}
