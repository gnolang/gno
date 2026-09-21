package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// entercrossingMachine builds a machine whose frame stack holds `depth` call
// frames, each interleaved with a non-call frame so the walk also exercises
// skipping them. m.Realm and every LastRealm are the same non-nil realm, so
// the implicit-switch gate is satisfied rather than vacuously true. The
// DEEPEST call frame (m.Frames[0]) is the one flagged, forcing a full walk.
func entercrossingMachine(depth int, budget int64, withCross, didCrossing bool) (*Machine, *Frame) {
	rlm := &Realm{Path: "gno.land/r/test/realm_a"}
	m := &Machine{
		Package:  &PackageValue{PkgPath: "gno.land/r/test/realm_a"},
		Realm:    rlm,
		GasMeter: types.NewGasMeter(budget),
	}
	for range depth {
		m.Frames = append(m.Frames,
			Frame{Func: &FuncValue{}, LastRealm: rlm},
			Frame{},
		)
	}
	m.Frames[0].WithCross = withCross
	m.Frames[0].DidCrossing = didCrossing
	return m, m.PeekCallFrame(1)
}

// TestDoOpEnterCrossingGasTotal pins the linear schedule: the walk charges
// OpCPUSlopeEnterCrossing per call frame visited, counting one extra virtual
// step for the faux deployer frame. Non-call frames do not count.
func TestDoOpEnterCrossingGasTotal(t *testing.T) {
	const slope = int64(OpCPUSlopeEnterCrossing)
	for _, tc := range []struct {
		name        string
		depth       int
		withCross   bool
		didCrossing bool
		want        int64
	}{
		{"first call", 1, true, false, 1 * slope},
		{"depth 2", 2, true, false, 2 * slope},
		{"depth 12", 12, true, false, 12 * slope},
		{"deep crossing", 100, true, false, 100 * slope},
		{"prior crossing", 100, false, true, 100 * slope},
		// No marked ancestor: the walk runs off the bottom and pays one
		// extra step for discovering the faux deployer frame.
		{"faux frame", 100, false, false, 101 * slope},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, fr1 := entercrossingMachine(tc.depth, tc.want, tc.withCross, tc.didCrossing)
			require.NotPanics(t, m.doOpEnterCrossing)
			require.Equal(t, tc.want*GasFactorCPU, m.GasMeter.GasConsumed())
			require.Equal(t, tc.want, m.Cycles)
			require.True(t, fr1.DidCrossing)
		})
	}
}

// TestDoOpEnterCrossingOutOfGas checks that a budget one gas short of the
// walk's cost aborts at the accept exit without marking the frame.
func TestDoOpEnterCrossingOutOfGas(t *testing.T) {
	const depth = 100
	want := int64(depth) * OpCPUSlopeEnterCrossing
	m, fr1 := entercrossingMachine(depth, want*GasFactorCPU-1, true, false)

	require.PanicsWithValue(t, types.OutOfGasError{Descriptor: "CPUCycles"}, m.doOpEnterCrossing)
	// The meter records the charge that broke the budget; Cycles is only
	// advanced after ConsumeGas returns, so it stays at zero.
	require.Equal(t, want*GasFactorCPU, m.GasMeter.GasConsumed())
	require.Equal(t, int64(0), m.Cycles)
	require.False(t, fr1.DidCrossing)
}

// TestDoOpEnterCrossingRealmMismatchChargesNoGas covers the implicit-switch
// gate: reaching a call frame whose LastRealm differs from the current realm
// with no cross(fn)(...) ancestor panics, and -- because the charge lands only
// at the accept exits -- the rejected walk costs no gas. Without this the gate
// branch has no coverage anywhere in the tree.
func TestDoOpEnterCrossingRealmMismatchChargesNoGas(t *testing.T) {
	rlm := &Realm{Path: "gno.land/r/test/realm_a"}
	other := &Realm{Path: "gno.land/r/test/realm_b"}
	m := &Machine{
		Package:  &PackageValue{PkgPath: "gno.land/r/test/realm_a"},
		Realm:    rlm,
		GasMeter: types.NewGasMeter(1 << 30),
	}
	// Deepest frame crossed from a different realm; top frame is clean, so
	// the walk accepts step 1 and rejects at step 2.
	m.Frames = append(m.Frames,
		Frame{Func: &FuncValue{}, LastRealm: other},
		Frame{Func: &FuncValue{}, LastRealm: rlm},
	)
	fr1 := m.PeekCallFrame(1)

	require.PanicsWithValue(t,
		"crossing could not find corresponding cross(fn)(...) call",
		m.doOpEnterCrossing)
	require.Equal(t, int64(0), m.GasMeter.GasConsumed())
	require.Equal(t, int64(0), m.Cycles)
	require.False(t, fr1.DidCrossing)
}
