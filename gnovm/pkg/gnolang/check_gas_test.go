package gnolang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// TestPreprocessGas_CheckWalks pins that the checker walks bill the
// preprocess meter: a function with n local declarations is walked once
// per node by checkUnused, so its charge grows with n.
func TestPreprocessGas_CheckWalks(t *testing.T) {
	charge := func(locals int) int64 {
		gm := stypes.NewGasMeter(1_000_000_000)
		st, _ := newPreprocessAllocTestStore(t, 256*1024*1024, gm)
		t.Cleanup(func() { st.SetPreprocessAllocator(nil) })
		m := NewMachineWithOptions(MachineOptions{
			PkgPath: "gno.land/r/test/checkgas",
			Store:   st,
		})
		defer m.Release()
		var b strings.Builder
		b.WriteString("package checkgas\n\nfunc F() {\n")
		for i := range locals {
			fmt.Fprintf(&b, "\tv%d := %d\n\t_ = v%d\n", i, i, i)
		}
		b.WriteString("}\n")
		panicked, val := runMemPackageRecover(m, &std.MemPackage{
			Type:  MPUserProd,
			Name:  "checkgas",
			Path:  "gno.land/r/test/checkgas",
			Files: []*std.MemFile{{Name: "a.gno", Body: b.String()}},
		})
		require.False(t, panicked, "got: %v", val)
		return gm.GasConsumed()
	}
	small, large := charge(10), charge(100)
	// At least one checker node per added declaration, on top of whatever
	// else scales with the statement count.
	require.GreaterOrEqual(t, large-small, int64(90)*OpCPUSlopeCheckNode*GasFactorCPU,
		"check walks must bill per declaration: %d -> %d", small, large)
}
