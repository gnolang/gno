package bitarray

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/random"
)

func randBitArray(bits int) (*BitArray, []byte) {
	src := random.RandBytes((bits + 7) / 8)
	bA := NewBitArray(bits)
	for i := range src {
		for j := range 8 {
			if i*8+j >= bits {
				return bA, src
			}
			setBit := src[i]&(1<<uint(j)) > 0
			bA.SetIndex(i*8+j, setBit)
		}
	}
	return bA, src
}

func TestAnd(t *testing.T) {
	t.Parallel()

	bA1, _ := randBitArray(51)
	bA2, _ := randBitArray(31)
	bA3 := bA1.And(bA2)

	var bNil *BitArray
	require.Equal(t, bNil.And(bA1), (*BitArray)(nil))
	require.Equal(t, bA1.And(nil), (*BitArray)(nil))
	require.Equal(t, bNil.And(nil), (*BitArray)(nil))

	if bA3.Bits != 31 {
		t.Error("Expected min bits", bA3.Bits)
	}
	if len(bA3.Elems) != len(bA2.Elems) {
		t.Error("Expected min elems length")
	}
	for i := range bA3.Bits {
		expected := bA1.GetIndex(i) && bA2.GetIndex(i)
		if bA3.GetIndex(i) != expected {
			t.Error("Wrong bit from bA3", i, bA1.GetIndex(i), bA2.GetIndex(i), bA3.GetIndex(i))
		}
	}
}

func TestOr(t *testing.T) {
	t.Parallel()

	bA1, _ := randBitArray(51)
	bA2, _ := randBitArray(31)
	bA3 := bA1.Or(bA2)

	bNil := (*BitArray)(nil)
	require.Equal(t, bNil.Or(bA1), bA1)
	require.Equal(t, bA1.Or(nil), bA1)
	require.Equal(t, bNil.Or(nil), (*BitArray)(nil))

	if bA3.Bits != 51 {
		t.Error("Expected max bits")
	}
	if len(bA3.Elems) != len(bA1.Elems) {
		t.Error("Expected max elems length")
	}
	for i := range bA3.Bits {
		expected := bA1.GetIndex(i) || bA2.GetIndex(i)
		if bA3.GetIndex(i) != expected {
			t.Error("Wrong bit from bA3", i, bA1.GetIndex(i), bA2.GetIndex(i), bA3.GetIndex(i))
		}
	}
}

func TestSub(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		initBA        string
		subtractingBA string
		expectedBA    string
	}{
		{`null`, `null`, `null`},
		{`"x"`, `null`, `null`},
		{`null`, `"x"`, `null`},
		{`"x"`, `"x"`, `"_"`},
		{`"xxxxxx"`, `"x_x_x_"`, `"_x_x_x"`},
		{`"x_x_x_"`, `"xxxxxx"`, `"______"`},
		{`"xxxxxx"`, `"x_x_x_xxxx"`, `"_x_x_x"`},
		{`"x_x_x_xxxx"`, `"xxxxxx"`, `"______xxxx"`},
		{`"xxxxxxxxxx"`, `"x_x_x_"`, `"_x_x_xxxxx"`},
		{`"x_x_x_"`, `"xxxxxxxxxx"`, `"______"`},
	}
	for _, tc := range testCases {
		var bA *BitArray
		err := json.Unmarshal([]byte(tc.initBA), &bA)
		require.Nil(t, err)

		var o *BitArray
		err = json.Unmarshal([]byte(tc.subtractingBA), &o)
		require.Nil(t, err)

		got, _ := json.Marshal(bA.Sub(o))
		require.Equal(t, tc.expectedBA, string(got), "%s minus %s doesn't equal %s", tc.initBA, tc.subtractingBA, tc.expectedBA)
	}
}

func TestPickRandom(t *testing.T) {
	t.Parallel()

	empty16Bits := "________________"
	empty64Bits := empty16Bits + empty16Bits + empty16Bits + empty16Bits
	testCases := []struct {
		bA string
		ok bool
	}{
		{`null`, false},
		{`"x"`, true},
		{`"` + empty16Bits + `"`, false},
		{`"x` + empty16Bits + `"`, true},
		{`"` + empty16Bits + `x"`, true},
		{`"x` + empty16Bits + `x"`, true},
		{`"` + empty64Bits + `"`, false},
		{`"x` + empty64Bits + `"`, true},
		{`"` + empty64Bits + `x"`, true},
		{`"x` + empty64Bits + `x"`, true},
	}
	for _, tc := range testCases {
		var bitArr *BitArray
		err := json.Unmarshal([]byte(tc.bA), &bitArr)
		require.NoError(t, err)
		_, ok := bitArr.PickRandom()
		require.Equal(t, tc.ok, ok, "PickRandom got an unexpected result on input %s", tc.bA)
	}
}

func TestBytes(t *testing.T) {
	t.Parallel()

	bA := NewBitArray(4)
	bA.SetIndex(0, true)
	check := func(bA *BitArray, bz []byte) {
		if !bytes.Equal(bA.Bytes(), bz) {
			panic(fmt.Sprintf("Expected %X but got %X", bz, bA.Bytes()))
		}
	}
	check(bA, []byte{0x01})
	bA.SetIndex(3, true)
	check(bA, []byte{0x09})

	bA = NewBitArray(9)
	check(bA, []byte{0x00, 0x00})
	bA.SetIndex(7, true)
	check(bA, []byte{0x80, 0x00})
	bA.SetIndex(8, true)
	check(bA, []byte{0x80, 0x01})

	bA = NewBitArray(16)
	check(bA, []byte{0x00, 0x00})
	bA.SetIndex(7, true)
	check(bA, []byte{0x80, 0x00})
	bA.SetIndex(8, true)
	check(bA, []byte{0x80, 0x01})
	bA.SetIndex(9, true)
	check(bA, []byte{0x80, 0x03})
}

func TestEmptyFull(t *testing.T) {
	t.Parallel()

	ns := []int{47, 123}
	for _, n := range ns {
		bA := NewBitArray(n)
		if !bA.IsEmpty() {
			t.Fatal("Expected bit array to be empty")
		}
		for i := range n {
			bA.SetIndex(i, true)
		}
		if !bA.IsFull() {
			t.Fatal("Expected bit array to be full")
		}
	}
}

func TestUpdateNeverPanics(t *testing.T) {
	t.Parallel()

	newRandBitArray := func(n int) *BitArray {
		ba, _ := randBitArray(n)
		return ba
	}
	pairs := []struct {
		a, b *BitArray
	}{
		{nil, nil},
		{newRandBitArray(10), newRandBitArray(12)},
		{newRandBitArray(23), newRandBitArray(23)},
		{newRandBitArray(37), nil},
		{nil, NewBitArray(10)},
	}

	for _, pair := range pairs {
		a, b := pair.a, pair.b
		a.Update(b)
		b.Update(a)
	}
}

func TestNewBitArrayNeverCrashesOnNegatives(t *testing.T) {
	t.Parallel()

	bitList := []int{-127, -128, -1 << 31}
	for _, bits := range bitList {
		_ = NewBitArray(bits)
	}
}

func TestJSONMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	bA1 := NewBitArray(0)

	bA2 := NewBitArray(1)

	bA3 := NewBitArray(1)
	bA3.SetIndex(0, true)

	bA4 := NewBitArray(5)
	bA4.SetIndex(0, true)
	bA4.SetIndex(1, true)

	testCases := []struct {
		bA           *BitArray
		marshalledBA string
	}{
		{nil, `null`},
		{bA1, `null`},
		{bA2, `"_"`},
		{bA3, `"x"`},
		{bA4, `"xx___"`},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.bA.String(), func(t *testing.T) {
			t.Parallel()

			bz, err := json.Marshal(tc.bA)
			require.NoError(t, err)

			assert.Equal(t, tc.marshalledBA, string(bz))

			var unmarshalledBA *BitArray
			err = json.Unmarshal(bz, &unmarshalledBA)
			require.NoError(t, err)

			if tc.bA == nil {
				require.Nil(t, unmarshalledBA)
			} else {
				require.NotNil(t, unmarshalledBA)
				assert.EqualValues(t, tc.bA.Bits, unmarshalledBA.Bits)
				if assert.EqualValues(t, tc.bA.String(), unmarshalledBA.String()) {
					assert.EqualValues(t, tc.bA.Elems, unmarshalledBA.Elems)
				}
			}
		})
	}
}

// Reproducer/regression test for json.Unmarshal crashing when no bits are passed into the JSON.
func TestUnmarshalJSONDoesntCrashOnZeroBits(t *testing.T) {
	t.Parallel()

	type indexCorpus struct {
		BitArray *BitArray `json:"ba"`
		Index    int       `json:"i"`
	}

	ic := new(indexCorpus)
	blob := []byte(`{"BA":""}`)
	err := json.Unmarshal(blob, ic)
	require.NoError(t, err)
	require.Equal(t, ic.BitArray, &BitArray{Bits: 0, Elems: nil})
}

func TestBitArrayValidateBasic(t *testing.T) {
	testCases := []struct {
		name    string
		bA1     *BitArray
		expPass bool
	}{
		{"valid empty", &BitArray{}, true},
		{"valid explicit 0 bits nil elements", &BitArray{Bits: 0, Elems: nil}, true},
		{"valid explicit 0 bits 0 len elements", &BitArray{Bits: 0, Elems: make([]uint64, 0)}, true},
		{"valid nil", nil, true},
		{"valid with elements", NewBitArray(10), true},
		{"more elements than bits specifies", &BitArray{Bits: 0, Elems: make([]uint64, 5)}, false},
		{"less elements than bits specifies", &BitArray{Bits: 200, Elems: make([]uint64, 1)}, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.bA1.ValidateBasic()
			require.Equal(t, err == nil, tc.expPass)
		})
	}
}
