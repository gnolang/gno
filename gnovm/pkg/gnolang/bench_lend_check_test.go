package gnolang

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// The BenchmarkLend programs make the loans they claim to measure: peak
// live loans per program, and none left inside an escaping callee.
func TestLendBenchLends(t *testing.T) {
	store := lendBenchStore(t)
	run := func(pkgPath string, x Expr) (peak int, inEscape []int) {
		lendArgsHook = func(m *Machine) {
			peak = max(peak, len(m.lends))
			if fv := m.Frames[len(m.Frames)-1].Func; fv != nil && (fv.Name == "keep" || fv.Name == "keepAll") {
				inEscape = append(inEscape, len(m.lends))
			}
		}
		defer func() { lendArgsHook = nil }()
		m := NewMachineWithOptions(MachineOptions{PkgPath: pkgPath, Store: store, Output: io.Discard})
		defer m.Release()
		m.Eval(x)
		require.Empty(t, m.lends)
		return peak, inEscape
	}
	for _, tc := range []struct {
		name string
		x    Expr
		peak int
	}{
		{"plain", Call(X("Plain"), 3), 0},
		{"lit1", Call(X("Lit1"), 3), 1},
		{"lit4", Call(X("Lit4"), 3), 4},
		{"named", Call(X("Named"), 3), 0},
		{"callback", Call(X("CallN"), 3, 10), 1},
		{"nest8", Call(X("Nest"), 1, 8), 9},
		{"nest64", Call(X("Nest"), 1, 64), 65},
		{"handon", Call(X("Drop"), 3, Str("handon")), 1},
		{"lit", Call(X("Drop"), 3, Str("lit")), 1},
		{"slice", Call(X("Drop"), 3, Str("slice")), 1},
	} {
		peak, inEscape := run(lendBenchRealmPath, tc.x)
		require.Equal(t, tc.peak, peak, tc.name)
		if tc.name == "handon" || tc.name == "lit" || tc.name == "slice" {
			require.Len(t, inEscape, 3, tc.name)
		}
		for _, n := range inEscape {
			require.Zero(t, n, "%s: a loan outlived its escape", tc.name)
		}
	}
	peak, _ := run(lendBenchCallPath, Call(X("Iter"), Nx(".origin"), X("true")))
	require.GreaterOrEqual(t, peak, 2, "avl write")
}
