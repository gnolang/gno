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

// TestPreprocessGas_TypeWalks pins the type-shape walk: a struct
// declaration with n map fields is walked once per field by
// checkTypeShape, so its charge grows with n.
func TestPreprocessGas_TypeWalks(t *testing.T) {
	charge := func(fields int) int64 {
		gm := stypes.NewGasMeter(1_000_000_000)
		st, _ := newPreprocessAllocTestStore(t, 256*1024*1024, gm)
		t.Cleanup(func() { st.SetPreprocessAllocator(nil) })
		m := NewMachineWithOptions(MachineOptions{
			PkgPath: "gno.land/r/test/checkgas",
			Store:   st,
		})
		defer m.Release()
		var b strings.Builder
		b.WriteString("package checkgas\n\ntype T struct {\n")
		for i := range fields {
			fmt.Fprintf(&b, "\tf%d map[string]int\n", i)
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
	require.GreaterOrEqual(t, large-small, int64(90)*OpCPUSlopeCheckNode*GasFactorCPU,
		"type walks must bill per field: %d -> %d", small, large)
}

// TestPreprocessGas_PromotedMethodExpr pins the lowering of a promoted
// method expression: T3.sum (T3 embeds T0) builds and preprocesses a
// literal with a node per parameter, metered per op; T0.sum, the same
// bytes, resolves directly and builds nothing.
func TestPreprocessGas_PromotedMethodExpr(t *testing.T) {
	charge := func(recv string, params int) int64 {
		gm := stypes.NewGasMeter(1_000_000_000)
		st, _ := newPreprocessAllocTestStore(t, 256*1024*1024, gm)
		t.Cleanup(func() { st.SetPreprocessAllocator(nil) })
		m := NewMachineWithOptions(MachineOptions{
			PkgPath: "gno.land/r/test/megas",
			Store:   st,
		})
		defer m.Release()
		var b strings.Builder
		b.WriteString("package megas\n\ntype T0 struct{}\n\nfunc (T0) sum(")
		for i := range params {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "a%d int", i)
		}
		b.WriteString(") int { return 0 }\n\ntype T3 struct{ T0 }\n\nvar F = " + recv + ".sum\n")
		panicked, val := runMemPackageRecover(m, &std.MemPackage{
			Type:  MPUserProd,
			Name:  "megas",
			Path:  "gno.land/r/test/megas",
			Files: []*std.MemFile{{Name: "a.gno", Body: b.String()}},
		})
		require.False(t, panicked, "got: %v", val)
		return gm.GasConsumed()
	}
	small := charge("T3", 10) - charge("T0", 10)
	large := charge("T3", 40) - charge("T0", 40)
	require.GreaterOrEqual(t, large-small, int64(30)*OpCPUSlopeCheckNode*GasFactorCPU,
		"the lowering must be metered per parameter: %d -> %d", small, large)
}
