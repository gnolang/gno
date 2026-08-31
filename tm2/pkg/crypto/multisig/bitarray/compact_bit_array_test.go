package bitarray

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randCompactBitArray(bits int) (*CompactBitArray, []byte) {
	numBytes := (bits + 7) / 8
	src := random.RandBytes((bits + 7) / 8)
	bA := NewCompactBitArray(bits)

	for i := range numBytes - 1 {
		for j := range uint8(8) {
			bA.SetIndex(i*8+int(j), src[i]&(uint8(1)<<(8-j)) > 0)
		}
	}
	// Set remaining bits
	for i := uint8(0); i < 8-bA.ExtraBitsStored; i++ {
		bA.SetIndex(numBytes*8+int(i), src[numBytes-1]&(uint8(1)<<(8-i)) > 0)
	}
	return bA, src
}

func TestNewBitArrayNeverCrashesOnNegatives(t *testing.T) {
	t.Parallel()

	bitList := []int{-127, -128, -1 << 31}
	for _, bits := range bitList {
		bA := NewCompactBitArray(bits)
		require.Nil(t, bA)
	}
}

func TestJSONMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	bA1 := NewCompactBitArray(0)
	bA2 := NewCompactBitArray(1)

	bA3 := NewCompactBitArray(1)
	bA3.SetIndex(0, true)

	bA4 := NewCompactBitArray(5)
	bA4.SetIndex(0, true)
	bA4.SetIndex(1, true)

	bA5 := NewCompactBitArray(9)
	bA5.SetIndex(0, true)
	bA5.SetIndex(1, true)
	bA5.SetIndex(8, true)

	bA6 := NewCompactBitArray(16)
	bA6.SetIndex(0, true)
	bA6.SetIndex(1, true)
	bA6.SetIndex(8, false)
	bA6.SetIndex(15, true)

	testCases := []struct {
		bA           *CompactBitArray
		marshalledBA string
	}{
		{nil, `null`},
		{bA1, `null`},
		{bA2, `"_"`},
		{bA3, `"x"`},
		{bA4, `"xx___"`},
		{bA5, `"xx______x"`},
		{bA6, `"xx_____________x"`},
	}

	for _, tc := range testCases {
		t.Run(tc.bA.String(), func(t *testing.T) {
			t.Parallel()

			bz, err := json.Marshal(tc.bA)
			require.NoError(t, err)

			assert.Equal(t, tc.marshalledBA, string(bz))

			var unmarshalledBA *CompactBitArray
			err = json.Unmarshal(bz, &unmarshalledBA)
			require.NoError(t, err)

			if tc.bA == nil {
				require.Nil(t, unmarshalledBA)
			} else {
				require.NotNil(t, unmarshalledBA)
				assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				if assert.EqualValues(t, tc.bA.String(), unmarshalledBA.String()) {
					assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				}
			}
		})
	}
}

func TestCompactMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	bA1 := NewCompactBitArray(0)
	bA2 := NewCompactBitArray(1)

	bA3 := NewCompactBitArray(1)
	bA3.SetIndex(0, true)

	bA4 := NewCompactBitArray(5)
	bA4.SetIndex(0, true)
	bA4.SetIndex(1, true)

	bA5 := NewCompactBitArray(9)
	bA5.SetIndex(0, true)
	bA5.SetIndex(1, true)
	bA5.SetIndex(8, true)

	bA6 := NewCompactBitArray(16)
	bA6.SetIndex(0, true)
	bA6.SetIndex(1, true)
	bA6.SetIndex(8, false)
	bA6.SetIndex(15, true)

	testCases := []struct {
		bA           *CompactBitArray
		marshalledBA []byte
	}{
		{nil, []byte("null")},
		{bA1, []byte("null")},
		{bA2, []byte{byte(1), byte(0)}},
		{bA3, []byte{byte(1), byte(128)}},
		{bA4, []byte{byte(5), byte(192)}},
		{bA5, []byte{byte(9), byte(192), byte(128)}},
		{bA6, []byte{byte(16), byte(192), byte(1)}},
	}

	for _, tc := range testCases {
		t.Run(tc.bA.String(), func(t *testing.T) {
			t.Parallel()

			bz := tc.bA.CompactMarshal()

			assert.Equal(t, tc.marshalledBA, bz)

			unmarshalledBA, err := CompactUnmarshal(bz)
			require.NoError(t, err)
			if tc.bA == nil {
				require.Nil(t, unmarshalledBA)
			} else {
				require.NotNil(t, unmarshalledBA)
				assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				if assert.EqualValues(t, tc.bA.String(), unmarshalledBA.String()) {
					assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				}
			}
		})
	}
}

func TestCompactBitArrayNumOfTrueBitsBefore(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		marshalledBA   string
		bAIndex        []int
		trueValueIndex []int
	}{
		{`"_____"`, []int{0, 1, 2, 3, 4}, []int{0, 0, 0, 0, 0}},
		{`"x"`, []int{0}, []int{0}},
		{`"_x"`, []int{1}, []int{0}},
		{`"x___xxxx"`, []int{0, 4, 5, 6, 7}, []int{0, 1, 2, 3, 4}},
		{`"__x_xx_x__x_x___"`, []int{2, 4, 5, 7, 10, 12}, []int{0, 1, 2, 3, 4, 5}},
		{`"______________xx"`, []int{14, 15}, []int{0, 1}},
	}
	for tcIndex, tc := range testCases {
		t.Run(tc.marshalledBA, func(t *testing.T) {
			t.Parallel()

			var bA *CompactBitArray
			err := json.Unmarshal([]byte(tc.marshalledBA), &bA)
			require.NoError(t, err)

			for i := range tc.bAIndex {
				require.Equal(t, tc.trueValueIndex[i], bA.NumTrueBitsBefore(tc.bAIndex[i]), "tc %d, i %d", tcIndex, i)
			}
		})
	}
}

func TestCompactBitArrayGetSetIndex(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(100))
	numTests := 10
	numBitsPerArr := 100
	for range numTests {
		bits := r.Intn(1000)
		bA, _ := randCompactBitArray(bits)

		for range numBitsPerArr {
			copied := bA.Copy()
			index := r.Intn(bits)
			val := (r.Int63() % 2) == 0
			bA.SetIndex(index, val)
			require.Equal(t, val, bA.GetIndex(index), "bA.SetIndex(%d, %v) failed on bit array: %s", index, val, copied)
		}
	}
}

// A CompactBitArray is decoded straight out of an untrusted transaction
// signature, so neither of the two properties ValidateBasic establishes holds
// for free: Size() is computed from ExtraBitsStored without reference to how
// many bytes Elems actually holds, and the bits past Size() are never read.
func TestValidateBasic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		bA      *CompactBitArray
		wantErr string
	}{
		{"nil", nil, ""},
		{"empty", &CompactBitArray{}, ""},
		{"full final byte", &CompactBitArray{ExtraBitsStored: 0, Elems: []byte{0xff}}, ""},
		{"partial final byte", &CompactBitArray{ExtraBitsStored: 3, Elems: []byte{0b1110_0000}}, ""},
		{
			// Size() would be 8 here, the same as ExtraBitsStored 0 over the
			// same byte: a second encoding of one bit array.
			"extra bits stored a full byte",
			&CompactBitArray{ExtraBitsStored: 8, Elems: []byte{0x00}},
			"extra bits stored is 8, want less than 8",
		},
		{
			// Size() would be 9, over the single byte stored.
			"extra bits stored past the elements",
			&CompactBitArray{ExtraBitsStored: 9, Elems: []byte{0x00}},
			"extra bits stored is 9, want less than 8",
		},
		{
			// Size() would be 1-8 == -7.
			"extra bits stored with no elements",
			&CompactBitArray{ExtraBitsStored: 1},
			"1 extra bits stored with no elements",
		},
		{
			"bits set past the end",
			&CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0b1000_0001}},
			"bits set past the end of the array",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.bA.ValidateBasic()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

// Everything NewCompactBitArray and SetIndex produce validates, whatever the
// size: rejecting a bit array a signer can actually build would reject the
// signature carrying it.
func TestValidateBasicAcceptsConstructedArrays(t *testing.T) {
	t.Parallel()

	for bits := 1; bits <= 64; bits++ {
		bA := NewCompactBitArray(bits)
		for i := range bits {
			bA.SetIndex(i, i%3 == 0)
		}
		require.NoError(t, bA.ValidateBasic(), "%d bits", bits)
	}
}

// The invariant the ExtraBitsStored clauses exist for: an array that validates
// can be walked all the way to Size() without GetIndex indexing past Elems.
func TestValidateBasicBoundsTheWalk(t *testing.T) {
	t.Parallel()

	for extra := range byte(16) {
		for elems := range 3 {
			bA := &CompactBitArray{ExtraBitsStored: extra, Elems: make([]byte, elems)}
			if bA.ValidateBasic() != nil {
				continue
			}
			require.NotPanics(t, func() { bA.NumTrueBitsBefore(bA.Size()) },
				"extra bits stored %d over %d elements reports size %d", extra, elems, bA.Size())
		}
	}
}
