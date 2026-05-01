// Package aminotest provides test helpers for verifying amino codec
// correctness across both the reflect-based codec and the genproto2
// generated codec. Import from _test.go files in any package whose types
// are registered with amino.RegisterGenproto2Type.
package aminotest

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// AssertCodecParity verifies that v round-trips identically through both
// the reflect codec (MarshalReflect / UnmarshalReflect) and the genproto2
// fast path (MarshalBinary2 / UnmarshalBinary2), with the two codecs
// agreeing on every wire byte and every decoded value.
//
// Asserted invariants:
//
//  1. Encoder parity: MarshalReflect(v) == MarshalBinary2(v).
//  2. Size invariant: SizeBinary2(v) == len(MarshalBinary2(v)).
//  3. Cross-decoder parity: UnmarshalReflect(bz) and UnmarshalBinary2(bz)
//     produce deeply-equal values.
//  4. Roundtrip fidelity: both decoded values reflect.DeepEqual to v.
//     This is strict — test inputs must avoid touching memoized caches
//     (e.g., don't call Commit.Hash() or MakeBlock() on the input, since
//     those populate unexported fields that won't be set on a freshly
//     decoded value and would spuriously fail DeepEqual).
//
// v must be a non-nil pointer to a value whose type implements
// amino.PBMessager2 (i.e., a type registered with genproto2).
func AssertCodecParity(t *testing.T, cdc *amino.Codec, v any) {
	t.Helper()

	rv := reflect.ValueOf(v)
	require.Equal(t, reflect.Ptr, rv.Kind(), "v must be a pointer, got %T", v)
	require.False(t, rv.IsNil(), "v must be non-nil")

	pbm, ok := v.(amino.PBMessager2)
	require.True(t, ok,
		"v (%T) must implement amino.PBMessager2; only genproto2-registered types are supported", v)

	// (1) Encoder parity.
	bzReflect, err := cdc.MarshalReflect(v)
	require.NoError(t, err, "MarshalReflect(%T)", v)
	bzBinary2, err := cdc.MarshalBinary2(pbm)
	require.NoError(t, err, "MarshalBinary2(%T)", v)
	require.Equal(t, bzReflect, bzBinary2,
		"encoder parity: MarshalReflect and MarshalBinary2 produced different bytes for %T", v)

	// (2) Size invariant.
	size, err := pbm.SizeBinary2(cdc)
	require.NoError(t, err, "SizeBinary2(%T)", v)
	require.Equal(t, len(bzBinary2), size,
		"size invariant: SizeBinary2(%d) != len(MarshalBinary2)(%d) for %T", size, len(bzBinary2), v)

	// Decode via reflect into fresh1.
	rt := rv.Type().Elem()
	fresh1Ptr := reflect.New(rt)
	require.NoError(t, cdc.UnmarshalReflect(bzReflect, fresh1Ptr.Interface()),
		"UnmarshalReflect into %T", fresh1Ptr.Interface())

	// Decode via genproto2 into fresh2.
	fresh2Ptr := reflect.New(rt)
	fresh2PBM, ok := fresh2Ptr.Interface().(amino.PBMessager2)
	require.True(t, ok, "fresh pointer %T must implement amino.PBMessager2", fresh2Ptr.Interface())
	require.NoError(t, fresh2PBM.UnmarshalBinary2(cdc, bzBinary2, 0),
		"UnmarshalBinary2 into %T", fresh2Ptr.Interface())

	fresh1Val := fresh1Ptr.Elem().Interface()
	fresh2Val := fresh2Ptr.Elem().Interface()
	origVal := rv.Elem().Interface()

	// (3) Cross-decoder parity.
	require.Equal(t, fresh1Val, fresh2Val,
		"cross-decoder parity: UnmarshalReflect and UnmarshalBinary2 produced different values for %T", v)

	// (4) Roundtrip fidelity.
	require.Equal(t, origVal, fresh1Val,
		"roundtrip fidelity: decoded value differs from original for %T", v)
}
