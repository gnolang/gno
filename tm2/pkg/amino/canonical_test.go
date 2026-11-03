package amino_test

import (
	"slices"
	"sync/atomic"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type canonicalInner struct {
	N int64
}

type canonicalOuter struct {
	A int64
	B string
	C string
	P *canonicalInner
}

func TestUnmarshalCanonical(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	value := canonicalOuter{A: 1, B: "x"}
	canonical, err := cdc.Marshal(value)
	require.NoError(t, err)
	// field 1 zigzag varint 1, field 2 one-byte string; C and P are omitted.
	require.Equal(t, []byte{0x08, 0x02, 0x12, 0x01, 'x'}, canonical)

	var out canonicalOuter
	require.NoError(t, cdc.UnmarshalCanonical(canonical, &out))
	assert.Equal(t, value, out)

	// A non-nil pointer to a zero struct is written as `tag 0x00` and must
	// survive the round trip: that encoding is canonical for that value.
	withPtr := canonicalOuter{A: 1, P: &canonicalInner{}}
	bz, err := cdc.Marshal(withPtr)
	require.NoError(t, err)
	require.Equal(t, []byte{0x08, 0x02, 0x22, 0x00}, bz)
	out = canonicalOuter{}
	require.NoError(t, cdc.UnmarshalCanonical(bz, &out))
	assert.Equal(t, withPtr, out)

	// Every mutation below decodes, with plain Unmarshal, to the same value
	// as canonical: that is the permissiveness UnmarshalCanonical removes.
	mutations := map[string][]byte{
		"default string field written out": {0x08, 0x02, 0x12, 0x01, 'x', 0x1a, 0x00},
		"non-minimal varint tag":           {0x88, 0x00, 0x02, 0x12, 0x01, 'x'},
		"non-minimal varint value":         {0x08, 0x82, 0x00, 0x12, 0x01, 'x'},
		"non-minimal length prefix":        {0x08, 0x02, 0x12, 0x81, 0x00, 'x'},
	}
	for name, bz := range mutations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var lenient canonicalOuter
			require.NoError(t, cdc.Unmarshal(bz, &lenient))
			assert.Equal(t, value, lenient)

			var strict canonicalOuter
			err := cdc.UnmarshalCanonical(bz, &strict)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "non-canonical encoding")
		})
	}

	// Malformed input fails as it does with Unmarshal.
	assert.Error(t, cdc.UnmarshalCanonical([]byte{0x08}, &out))
}

// A local struct only exercises the reflect encoder and decoder. Generated
// types take the genproto2 path both ways, as std.Tx does in BaseApp.runTx.
func TestUnmarshalCanonicalGenproto2(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()
	stats := cdc.GetStats()

	gen, err := cdc.Marshal(tests.PrimitivesStruct{Int8: 1})
	require.NoError(t, err)
	// Field 1 zigzag varint 1, then field 18, the zero time.Time, as an
	// 11-byte message; the bytes padded below are all single-byte varints.
	require.Equal(t, []byte{0x08, 0x02, 0x92, 0x01, 0x0b}, gen[:5])

	encodes, decodes := atomic.LoadInt64(&stats.Genproto2Encodes), atomic.LoadInt64(&stats.Genproto2Decodes)
	var g tests.PrimitivesStruct
	require.NoError(t, cdc.UnmarshalCanonical(gen, &g))
	assert.Equal(t, int8(1), g.Int8)
	assert.Equal(t, encodes+1, atomic.LoadInt64(&stats.Genproto2Encodes), "re-encoding took the genproto2 path")
	assert.Equal(t, decodes+1, atomic.LoadInt64(&stats.Genproto2Decodes), "decoding took the genproto2 path")

	mutations := map[string][]byte{
		"non-minimal varint tag":            slices.Concat([]byte{gen[0] | 0x80, 0x00}, gen[1:]),
		"non-minimal varint value":          slices.Concat(gen[:1], []byte{gen[1] | 0x80, 0x00}, gen[2:]),
		"non-minimal nested message length": slices.Concat(gen[:4], []byte{gen[4] | 0x80, 0x00}, gen[5:]),
	}
	for name, bz := range mutations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var lenient tests.PrimitivesStruct
			require.NoError(t, cdc.Unmarshal(bz, &lenient))
			assert.Equal(t, int8(1), lenient.Int8)

			var strict tests.PrimitivesStruct
			require.ErrorContains(t, cdc.UnmarshalCanonical(bz, &strict), "non-canonical encoding")
		})
	}
}

// Unmarshal takes *T and **T, and its reflect path decodes through deeper
// pointers; Marshal panics on more than one level. UnmarshalCanonical must
// return an error, not decode and then panic.
func TestUnmarshalCanonicalPointers(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	value := canonicalOuter{A: 1, B: "x"}
	bz, err := cdc.Marshal(value)
	require.NoError(t, err)

	var p *canonicalOuter
	require.NoError(t, cdc.UnmarshalCanonical(bz, &p))
	require.NotNil(t, p)
	assert.Equal(t, value, *p)

	var pp **canonicalOuter
	err = cdc.UnmarshalCanonical(bz, &pp)
	require.ErrorContains(t, err, "nested pointers not allowed")
	assert.Nil(t, pp, "refused before decoding")

	assert.ErrorIs(t, cdc.UnmarshalCanonical(bz, value), amino.ErrNoPointer)
}
