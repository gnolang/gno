package amino_test

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
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
