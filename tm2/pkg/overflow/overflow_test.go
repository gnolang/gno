package overflow

import (
	"math"
	"math/big"
	"math/rand/v2"
	"testing"
)

func TestSigned(t *testing.T) {
	var (
		nr = big.NewInt(0)
		// returned when dividing by zero - not representable by int64,
		// has 0 remainder.
		divByZero = big.NewInt(0).Lsh(big.NewInt(1), 64)
		// throwaway values.
		sink = big.NewInt(0)
		// used to set big i and j.
		bi, bj = big.NewInt(0), big.NewInt(0)
	)
	tt := []struct {
		opName string
		fn8    func(a, b int8) (int8, bool)
		fn16   func(a, b int16) (int16, bool)
		fn32   func(a, b int32) (int32, bool)
		fn64   func(a, b int64) (int64, bool)
		bigFn  func(x, y *big.Int) *big.Int
	}{
		{
			"Add",
			Add[int8], Add[int16], Add[int32], Add[int64],
			nr.Add,
		},
		{
			"Sub",
			Sub[int8], Sub[int16], Sub[int32], Sub[int64],
			nr.Sub,
		},
		{
			"Mul",
			Mul[int8], Mul[int16], Mul[int32], Mul[int64],
			nr.Mul,
		},
		{
			"Div",
			Div[int8], Div[int16], Div[int32], Div[int64],
			func(x, y *big.Int) *big.Int {
				if y.Int64() == 0 {
					return divByZero
				}
				result, _ := nr.QuoRem(x, y, sink)
				return result
			},
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
			for i := range int64(1<<16 - 1) {
				// Construct i and j: the high byte is set, the later 7 bytes
				// are random.
				i, j := mk(i>>8), mk(i&0xFF)

				// For each bit size, test the function, ignoring random LSB's
				// as necessary. The value is cross-checked against using the
				// equivalent function of math/big.
				{
					i8, j8 := int8(i>>56), int8(j>>56)
					r8, ok := tc.fn8(i8, j8)
					errors += checkResult(
						t, tc.opName,
						i8, j8, r8, ok,
						tc.bigFn(bi.SetInt64(int64(i8)), bj.SetInt64(int64(j8))),
					)
				}
				{
					i16, j16 := int16(i>>48), int16(j>>48)
					r16, ok := tc.fn16(i16, j16)
					errors += checkResult(
						t, tc.opName,
						i16, j16, r16, ok,
						tc.bigFn(bi.SetInt64(int64(i16)), bj.SetInt64(int64(j16))),
					)
				}
				{
					i32, j32 := int32(i>>32), int32(j>>32)
					r32, ok := tc.fn32(i32, j32)
					errors += checkResult(
						t, tc.opName,
						i32, j32, r32, ok,
						tc.bigFn(bi.SetInt64(int64(i32)), bj.SetInt64(int64(j32))),
					)
				}
				{
					r64, ok := tc.fn64(i, j)
					errors += checkResult(
						t, tc.opName,
						i, j, r64, ok,
						tc.bigFn(bi.SetInt64(i), bj.SetInt64(j)),
					)
				}

				if errors > 100 {
					t.Error("too many errors")
					return
				}
			}
		})
	}
}

//nolint:thelper // t.Helper significantly slows this down, and the errors already contain enough information.
func checkResult[T int8 | int16 | int32 | int64](
	t *testing.T,
	op string,
	i, j, r T,
	ok bool,
	bigRes *big.Int,
) int {
	resI64 := bigRes.Int64()
	expR := T(resI64)
	expOk := bigRes.IsInt64() && int64(T(resI64)) == resI64
	if r != expR {
		t.Errorf("%T: %d %s %d: expected %d got %d", T(0), i, op, j, expR, r)
		return 1
	}
	if ok != expOk {
		t.Errorf("%T: %d %s %d: expected %t got %t", T(0), i, op, j, expOk, ok)
		return 1
	}
	return 0
}

func TestUnsigned(t *testing.T) {
	var (
		nr = big.NewInt(0)
		// returned when dividing by zero - not representable by uint64,
		// has 0 remainder.
		divByZero = big.NewInt(0).Lsh(big.NewInt(1), 65)
		// throwaway values.
		sink = big.NewInt(0)
		// used to set big i and j.
		bi, bj = big.NewInt(0), big.NewInt(0)
	)
	tt := []struct {
		opName string
		fn8    func(a, b uint8) (uint8, bool)
		fn16   func(a, b uint16) (uint16, bool)
		fn32   func(a, b uint32) (uint32, bool)
		fn64   func(a, b uint64) (uint64, bool)
		bigFn  func(x, y *big.Int) *big.Int
	}{
		{
			"Add",
			Add[uint8], Add[uint16], Add[uint32], Add[uint64],
			nr.Add,
		},
		{
			"Sub",
			Sub[uint8], Sub[uint16], Sub[uint32], Sub[uint64],
			nr.Sub,
		},
		{
			"Mul",
			Mul[uint8], Mul[uint16], Mul[uint32], Mul[uint64],
			nr.Mul,
		},
		{
			"Div",
			Div[uint8], Div[uint16], Div[uint32], Div[uint64],
			func(x, y *big.Int) *big.Int {
				if y.Uint64() == 0 {
					return divByZero
				}
				result, _ := nr.QuoRem(x, y, sink)
				return result
			},
		},
	}
	mk := func(i uint64) uint64 {
		// Ensure to always have "edge case" values.
		switch i {
		case 0:
			return 0
		case 255:
			return 1<<64 - 1
		default:
			return i<<56 | rand.Uint64N(1<<56)
		}
	}
	for _, tc := range tt {
		t.Run(tc.opName, func(t *testing.T) {
			errors := 0
			for i := range uint64(1<<16 - 1) {
				// Construct i and j: the high byte is set, the later 7 bytes
				// are random.
				i, j := mk(i>>8), mk(i&0xFF)

				// For each bit size, test the function, ignoring random LSB's
				// as necessary. The value is cross-checked against using the
				// equivalent function of math/big.
				{
					i8, j8 := uint8(i>>56), uint8(j>>56)
					r8, ok := tc.fn8(i8, j8)
					errors += checkResultUnsigned(
						t, tc.opName,
						i8, j8, r8, ok,
						tc.bigFn(bi.SetUint64(uint64(i8)), bj.SetUint64(uint64(j8))),
					)
				}
				{
					i16, j16 := uint16(i>>48), uint16(j>>48)
					r16, ok := tc.fn16(i16, j16)
					errors += checkResultUnsigned(
						t, tc.opName,
						i16, j16, r16, ok,
						tc.bigFn(bi.SetUint64(uint64(i16)), bj.SetUint64(uint64(j16))),
					)
				}
				{
					i32, j32 := uint32(i>>32), uint32(j>>32)
					r32, ok := tc.fn32(i32, j32)
					errors += checkResultUnsigned(
						t, tc.opName,
						i32, j32, r32, ok,
						tc.bigFn(bi.SetUint64(uint64(i32)), bj.SetUint64(uint64(j32))),
					)
				}
				{
					r64, ok := tc.fn64(i, j)
					errors += checkResultUnsigned(
						t, tc.opName,
						i, j, r64, ok,
						tc.bigFn(bi.SetUint64(i), bj.SetUint64(j)),
					)
				}

				if errors > 100 {
					t.Error("too many errors")
					return
				}
			}
		})
	}
}

//nolint:thelper // t.Helper significantly slows this down, and the errors already contain enough information.
func checkResultUnsigned[T uint8 | uint16 | uint32 | uint64](
	t *testing.T,
	op string,
	i, j, r T,
	ok bool,
	bigRes *big.Int,
) int {
	resU64 := bigRes.Uint64()
	expR := T(resU64)
	if bigRes.Sign() == -1 {
		expR = 0 - expR
	}
	expOk := bigRes.IsUint64() && uint64(T(resU64)) == resU64
	if r != expR {
		t.Errorf("%T: %d %s %d: expected %d got %d", T(0), i, op, j, expR, r)
		return 1
	}
	if ok != expOk {
		t.Errorf("%T: %d %s %d: expected %t got %t", T(0), i, op, j, expOk, ok)
		return 1
	}
	return 0
}
