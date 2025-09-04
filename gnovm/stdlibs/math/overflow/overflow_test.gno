package overflow

import (
	"math"
	"math/rand"
	"testing"
)

func TestSigned(t *testing.T) {
	tt := []struct {
		opName string
		fn8    func(a, b int8) (int8, bool)
		fn16   func(a, b int16) (int16, bool)
		fn32   func(a, b int32) (int32, bool)
		fn64   func(a, b int64) (int64, bool)
	}{
		{
			"Add",
			Add8, Add16, Add32, Add64,
		},
		{
			"Sub",
			Sub8, Sub16, Sub32, Sub64,
		},
		{
			"Mul",
			Mul8, Mul16, Mul32, Mul64,
		},
		{
			"Div",
			Div8, Div16, Div32, Div64,
		},
	}
	mk := func(i int64) int64 {
		// Ensure to always have "edge case" values.
		switch i {
		case 0:
			return 0
		case 127:
			// biggest positive int
			return math.MaxInt64
		case 128:
			// smallest negative int
			return math.MinInt64
		case 255:
			// negative all 1's bits
			return -1
		default:
			return i<<56 | rand.Int64N(1<<56)
		}
	}
	for _, tc := range tt {
		t.Run(tc.opName, func(t *testing.T) {
			errors := 0
			for x := int64(0); x < (1<<16)-1; x++ {
				// Construct i and j: the high byte is set, the later 7 bytes
				// are random.
				i, j := mk(x>>8), mk(x&0xFF)

				// For each bit size, test the function, ignoring random LSB's
				// as necessary. The value is cross-checked against using the
				// equivalent in int64.
				{
					i8, j8 := int8(i>>56), int8(j>>56)
					r8, ok := tc.fn8(i8, j8)
					r64, ok64 := tc.fn64(i>>56, j>>56)
					errors += checkResult(
						t, tc.opName, 8,
						int64(i8), int64(j8),
						int64(r8), ok,
						r64, ok64,
					)
				}
				{
					i16, j16 := int16(i>>48), int16(j>>48)
					r16, ok := tc.fn16(i16, j16)
					r64, ok64 := tc.fn64(i>>48, j>>48)
					errors += checkResult(
						t, tc.opName, 16,
						int64(i16), int64(j16),
						int64(r16), ok,
						r64, ok64,
					)
				}
				{
					i32, j32 := int32(i>>32), int32(j>>32)
					r32, ok := tc.fn32(i32, j32)
					r64, ok64 := tc.fn64(i>>32, j>>32)
					errors += checkResult(
						t, tc.opName, 32,
						int64(i32), int64(j32),
						int64(r32), ok,
						r64, ok64,
					)
				}
				// fn64 is the baseline; no need to test it.

				if errors > 100 {
					t.Error("too many errors")
					return
				}
			}
		})
	}
}

func checkResult(
	t *testing.T,
	op string, bits uint, // name of op, like "Add", and the number of bits of the op.
	i, j int64, // added values
	r int64, ok bool, // result in the target bit size
	r64 int64, ok64 bool, // result using the 64-bit version
) int {
	mask := int64(1<<bits) - 1
	expR := r64 & mask
	if expR&(1<<(bits-1)) != 0 {
		// If it's negative, subtract 2^bits.
		// For example: int8(-32) is 0xE0 in binary, would be 0x00000000E0 as int64 (224).
		// We subtract 2^8, so this becomes 0xFFFFFFFFFFFFFFE0, ie. -32 also in int64.
		expR -= 1 << bits
	}
	expOk := ok64 && expR == r64
	if r != expR {
		t.Errorf("int%d: %d %s %d: expected %d got %d", bits, i, op, j, expR, r)
		return 1
	}
	if ok != expOk {
		t.Errorf("int%d: %d %s %d: expected %t got %t", bits, i, op, j, expOk, ok)
		return 1
	}
	return 0
}

func TestUnsigned(t *testing.T) {
	tt := []struct {
		opName string
		fn8    func(a, b uint8) (uint8, bool)
		fn16   func(a, b uint16) (uint16, bool)
		fn32   func(a, b uint32) (uint32, bool)
		fn64   func(a, b uint64) (uint64, bool)
	}{
		{
			"Add",
			Addu8, Addu16, Addu32, Addu64,
		},
		{
			"Sub",
			Subu8, Subu16, Subu32, Subu64,
		},
		{
			"Mul",
			Mulu8, Mulu16, Mulu32, Mulu64,
		},
		{
			"Div",
			Divu8, Divu16, Divu32, Divu64,
		},
	}
	mk := func(i uint64) uint64 {
		// Ensure to always have "edge case" values.
		switch i {
		case 0:
			return 0
		case 255:
			return (1 << 64) - 1
		default:
			return i<<56 | rand.Uint64N(1<<56)
		}
	}
	for _, tc := range tt {
		t.Run(tc.opName, func(t *testing.T) {
			errors := 0
			for x := uint64(0); x < (1<<16)-1; x++ {
				// Construct i and j: the high byte is set, the later 7 bytes
				// are random.
				i, j := mk(x>>8), mk(x&0xFF)

				// For each bit size, test the function, ignoring random LSB's
				// as necessary. The value is cross-checked against using the
				// equivalent in uint64.
				{
					i8, j8 := uint8(i>>56), uint8(j>>56)
					r8, ok := tc.fn8(i8, j8)
					r64, ok64 := tc.fn64(i>>56, j>>56)
					errors += checkResultUnsigned(
						t, tc.opName, 8,
						uint64(i8), uint64(j8),
						uint64(r8), ok,
						r64, ok64,
					)
				}
				{
					i16, j16 := uint16(i>>48), uint16(j>>48)
					r16, ok := tc.fn16(i16, j16)
					r64, ok64 := tc.fn64(i>>48, j>>48)
					errors += checkResultUnsigned(
						t, tc.opName, 16,
						uint64(i16), uint64(j16),
						uint64(r16), ok,
						r64, ok64,
					)
				}
				{
					i32, j32 := uint32(i>>32), uint32(j>>32)
					r32, ok := tc.fn32(i32, j32)
					r64, ok64 := tc.fn64(i>>32, j>>32)
					errors += checkResultUnsigned(
						t, tc.opName, 32,
						uint64(i32), uint64(j32),
						uint64(r32), ok,
						r64, ok64,
					)
				}
				// fn64 is the baseline; no need to test it.

				if errors > 100 {
					t.Error("too many errors")
					return
				}
			}
		})
	}
}

func checkResultUnsigned(
	t *testing.T,
	op string, bits uint, // name of op, like "Add", and the number of bits of the op.
	i, j uint64, // added values
	r uint64, ok bool, // result in the target bit size
	r64 uint64, ok64 bool, // result using the 64-bit version
) int {
	mask := uint64(1<<bits) - 1
	expR := r64 & mask
	expOk := ok64 && expR == r64
	if r != expR {
		t.Errorf("int%d: %d %s %d: expected %d got %d", bits, i, op, j, expR, r)
		return 1
	}
	if ok != expOk {
		t.Errorf("int%d: %d %s %d: expected %t got %t", bits, i, op, j, expOk, ok)
		return 1
	}
	return 0
}
