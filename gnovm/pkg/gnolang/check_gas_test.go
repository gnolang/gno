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

// TestPreprocessGas_AssertWalks pins the impossible-assertion walk: a type
// assertion and a type switch case against a type that embeds d levels
// deep cost more as d grows, so the embedding walk is metered.
func TestPreprocessGas_AssertWalks(t *testing.T) {
	charge := func(depth int, assert bool) int64 {
		gm := stypes.NewGasMeter(1_000_000_000)
		st, _ := newPreprocessAllocTestStore(t, 256*1024*1024, gm)
		t.Cleanup(func() { st.SetPreprocessAllocator(nil) })
		m := NewMachineWithOptions(MachineOptions{
			PkgPath: "gno.land/r/test/checkgas",
			Store:   st,
		})
		defer m.Release()
		var b strings.Builder
		b.WriteString("package checkgas\n\ntype I interface{ M() }\ntype E0 struct{}\n\nfunc (E0) M() {}\n")
		for i := 1; i <= depth; i++ {
			fmt.Fprintf(&b, "type E%d struct{ E%d }\n", i, i-1)
		}
		if assert {
			fmt.Fprintf(&b, "func f(i I) { _ = i.(E%d); switch i.(type) { case E%d: } }\n", depth, depth)
		} else {
			b.WriteString("func f(i I) { _ = i }\n")
		}
		panicked, val := runMemPackageRecover(m, &std.MemPackage{
			Type:  MPUserProd,
			Name:  "checkgas",
			Path:  "gno.land/r/test/checkgas",
			Files: []*std.MemFile{{Name: "a.gno", Body: b.String()}},
		})
		require.False(t, panicked, "got: %v", val)
		return gm.GasConsumed()
	}
	shallow := charge(1, true) - charge(1, false)
	deep := charge(8, true) - charge(8, false)
	require.GreaterOrEqual(t, deep-shallow, int64(7)*OpCPUSlopeEmbedExpand*GasFactorCPU,
		"assertion walks must bill per embedding level: %d -> %d", shallow, deep)
}
