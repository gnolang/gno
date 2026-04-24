package amino_test

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

// TestUnmarshalDoublePointer_Genproto2 verifies that
//
//	var p *T; amino.Unmarshal(bz, &p)
//
// takes the genproto2 fast path (not reflect), and produces
// the correct result.
func TestUnmarshalDoublePointer_Genproto2(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	orig := tests.PrimitivesStruct{
		Int8: 42, Int16: 1000, Str: "hello",
	}
	bz, err := cdc.Marshal(&orig)
	require.NoError(t, err)

	// Record genproto2 counter before.
	g2before := atomic.LoadInt64(&cdc.GetStats().Genproto2Decodes)

	// Unmarshal via **T pattern.
	var dst *tests.PrimitivesStruct
	err = cdc.Unmarshal(bz, &dst)
	require.NoError(t, err)

	// Verify correctness.
	require.NotNil(t, dst)
	assert.Equal(t, orig.Int8, dst.Int8)
	assert.Equal(t, orig.Int16, dst.Int16)
	assert.Equal(t, orig.Str, dst.Str)

	// Verify genproto2 path was taken (counter incremented by 1).
	g2after := atomic.LoadInt64(&cdc.GetStats().Genproto2Decodes)
	assert.Equal(t, g2before+1, g2after, "expected genproto2 path for **T unmarshal")
}

// TestUnmarshalDoublePointer_NonNilInner verifies that **T with a
// non-nil inner pointer decodes into the existing struct.
func TestUnmarshalDoublePointer_NonNilInner(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	orig := tests.PrimitivesStruct{Int8: 99, Str: "world"}
	bz, err := cdc.Marshal(&orig)
	require.NoError(t, err)

	existing := &tests.PrimitivesStruct{Int8: 1, Str: "old"}
	err = cdc.Unmarshal(bz, &existing)
	require.NoError(t, err)
	assert.Equal(t, int8(99), existing.Int8)
	assert.Equal(t, "world", existing.Str)
}

// TestUnmarshalDoublePointer_Roundtrip verifies full roundtrip
// equality through the **T pattern.
func TestUnmarshalDoublePointer_Roundtrip(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	orig := tests.PrimitivesStruct{
		Int8: 127, Int16: -32000, Int32: 100000,
		Str: "roundtrip", Byte: 0xFF,
	}
	bz, err := cdc.Marshal(&orig)
	require.NoError(t, err)

	var dst *tests.PrimitivesStruct
	err = cdc.Unmarshal(bz, &dst)
	require.NoError(t, err)
	require.NotNil(t, dst)

	// Compare key fields (full DeepEqual may differ on time zero-values).
	assert.Equal(t, orig.Int8, dst.Int8)
	assert.Equal(t, orig.Int16, dst.Int16)
	assert.Equal(t, orig.Int32, dst.Int32)
	assert.Equal(t, orig.Str, dst.Str)
	assert.Equal(t, orig.Byte, dst.Byte)
}

// TestMarshalBareValue_Genproto2 verifies that Marshal(val) where val
// is bare T (not *T) now hits the genproto2 fast path via PBMarshaler2.
func TestMarshalBareValue_Genproto2(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	orig := tests.PrimitivesStruct{Int8: 42, Str: "bare"}

	g2before := atomic.LoadInt64(&cdc.GetStats().Genproto2Encodes)

	// Pass bare value, not pointer.
	bz, err := cdc.Marshal(orig)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	g2after := atomic.LoadInt64(&cdc.GetStats().Genproto2Encodes)
	assert.Equal(t, g2before+1, g2after, "bare T should use genproto2 via PBMarshaler2")

	// Verify bytes match pointer-based marshal.
	bz2, err := cdc.Marshal(&orig)
	require.NoError(t, err)
	assert.Equal(t, bz, bz2, "bare T and *T should produce identical bytes")
}

// TestUnmarshalSinglePointer_StillWorks verifies that the normal
// *T pattern is unaffected by the **T peel.
func TestUnmarshalSinglePointer_StillWorks(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	orig := tests.PrimitivesStruct{Int8: 7}
	bz, err := cdc.Marshal(&orig)
	require.NoError(t, err)

	g2before := atomic.LoadInt64(&cdc.GetStats().Genproto2Decodes)

	var dst tests.PrimitivesStruct
	err = cdc.Unmarshal(bz, &dst)
	require.NoError(t, err)
	assert.Equal(t, int8(7), dst.Int8)

	g2after := atomic.LoadInt64(&cdc.GetStats().Genproto2Decodes)
	assert.Equal(t, g2before+1, g2after, "single pointer should also use genproto2")
}
