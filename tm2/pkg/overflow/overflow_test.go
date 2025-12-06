package overflow

import (
	"math"
	"math/big"
	"math/rand/v2"
	"testing"
)

func TestSignedInt(t *testing.T) {
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
					errors += checkResultSignedInt(
						t, tc.opName,
						i8, j8, r8, ok,
						tc.bigFn(bi.SetInt64(int64(i8)), bj.SetInt64(int64(j8))),
					)
				}
				{
					i16, j16 := int16(i>>48), int16(j>>48)
					r16, ok := tc.fn16(i16, j16)
					errors += checkResultSignedInt(
						t, tc.opName,
						i16, j16, r16, ok,
						tc.bigFn(bi.SetInt64(int64(i16)), bj.SetInt64(int64(j16))),
					)
				}
				{
					i32, j32 := int32(i>>32), int32(j>>32)
					r32, ok := tc.fn32(i32, j32)
					errors += checkResultSignedInt(
						t, tc.opName,
						i32, j32, r32, ok,
						tc.bigFn(bi.SetInt64(int64(i32)), bj.SetInt64(int64(j32))),
					)
				}
				{
					r64, ok := tc.fn64(i, j)
					errors += checkResultSignedInt(
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
func checkResultSignedInt[T int8 | int16 | int32 | int64](
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

func TestUnsignedInt(t *testing.T) {
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
					errors += checkResultUnsignedInt(
						t, tc.opName,
						i8, j8, r8, ok,
						tc.bigFn(bi.SetUint64(uint64(i8)), bj.SetUint64(uint64(j8))),
					)
				}
				{
					i16, j16 := uint16(i>>48), uint16(j>>48)
					r16, ok := tc.fn16(i16, j16)
					errors += checkResultUnsignedInt(
						t, tc.opName,
						i16, j16, r16, ok,
						tc.bigFn(bi.SetUint64(uint64(i16)), bj.SetUint64(uint64(j16))),
					)
				}
				{
					i32, j32 := uint32(i>>32), uint32(j>>32)
					r32, ok := tc.fn32(i32, j32)
					errors += checkResultUnsignedInt(
						t, tc.opName,
						i32, j32, r32, ok,
						tc.bigFn(bi.SetUint64(uint64(i32)), bj.SetUint64(uint64(j32))),
					)
				}
				{
					r64, ok := tc.fn64(i, j)
					errors += checkResultUnsignedInt(
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
func checkResultUnsignedInt[T uint8 | uint16 | uint32 | uint64](
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

func TestFloat(t *testing.T) {
	tt := []struct {
		opName string
		fn32   func(a, b float32) (float32, bool)
		fn64   func(a, b float64) (float64, bool)
		// For floats, we validate differently than integers:
		// the operation succeeds if result is finite (not Inf/NaN)
		shouldSucceed func(a, b, result float64) bool
	}{
		{
			"Add",
			Add[float32], Add[float64],
			func(a, b, result float64) bool {
				// Addition succeeds if result is finite
				return !math.IsInf(result, 0) && !math.IsNaN(result)
			},
		},
		{
			"Sub",
			Sub[float32], Sub[float64],
			func(a, b, result float64) bool {
				// Subtraction succeeds if result is finite
				return !math.IsInf(result, 0) && !math.IsNaN(result)
			},
		},
		{
			"Mul",
			Mul[float32], Mul[float64],
			func(a, b, result float64) bool {
				// Multiplication succeeds if result is finite
				return !math.IsInf(result, 0) && !math.IsNaN(result)
			},
		},
		{
			"Div",
			Div[float32], Div[float64],
			func(a, b, result float64) bool {
				// Division succeeds if divisor is non-zero and result is finite
				if b == 0 {
					return false
				}
				return !math.IsInf(result, 0) && !math.IsNaN(result)
			},
		},
	}

	// Edge case values for float testing
	mk := func(i int) float64 {
		switch i {
		case 0:
			return 0
		case 1:
			return 1
		case 2:
			return -1
		case 3:
			// Very small positive
			return 1e-38
		case 4:
			// Very small negative
			return -1e-38
		case 5:
			// Large positive (close to float32 max)
			return 3e38
		case 6:
			// Large negative
			return -3e38
		case 7:
			// Very large (close to float64 max)
			return 1e308
		case 8:
			// Very large negative
			return -1e308
		case 9:
			// Max float32
			return math.MaxFloat32
		case 10:
			// Max float64
			return math.MaxFloat64
		default:
			// Random value scaled by magnitude
			magnitude := float64(i % 100)
			return (rand.Float64()*2 - 1) * math.Pow(10, magnitude)
		}
	}

	for _, tc := range tt {
		t.Run(tc.opName, func(t *testing.T) {
			errors := 0
			// Test fewer iterations than integer tests since float validation is simpler
			for i := range 1000 {
				a, b := mk(i), mk((i*7)%256)

				// Test float32
				{
					a32, b32 := float32(a), float32(b)
					// Compute result using float32 precision
					var r32Float float32
					switch tc.opName {
					case "Add":
						r32Float = a32 + b32
					case "Sub":
						r32Float = a32 - b32
					case "Mul":
						r32Float = a32 * b32
					case "Div":
						if b32 != 0 {
							r32Float = a32 / b32
						}
					}

					r32, ok := tc.fn32(a32, b32)
					expOk := tc.shouldSucceed(float64(a32), float64(b32), float64(r32Float))

					if r32 != r32Float && !(math.IsNaN(float64(r32)) && math.IsNaN(float64(r32Float))) {
						t.Errorf("float32: %v %s %v: expected %v got %v", a32, tc.opName, b32, r32Float, r32)
						errors++
					}
					if ok != expOk {
						t.Errorf("float32: %v %s %v: expected ok=%t got ok=%t", a32, tc.opName, b32, expOk, ok)
						errors++
					}
				}

				// Test float64
				{
					var r64Float float64
					switch tc.opName {
					case "Add":
						r64Float = a + b
					case "Sub":
						r64Float = a - b
					case "Mul":
						r64Float = a * b
					case "Div":
						if b != 0 {
							r64Float = a / b
						}
					}

					r64, ok := tc.fn64(a, b)
					expOk := tc.shouldSucceed(a, b, r64Float)

					if r64 != r64Float && !(math.IsNaN(r64) && math.IsNaN(r64Float)) {
						t.Errorf("float64: %v %s %v: expected %v got %v", a, tc.opName, b, r64Float, r64)
						errors++
					}
					if ok != expOk {
						t.Errorf("float64: %v %s %v: expected ok=%t got ok=%t", a, tc.opName, b, expOk, ok)
						errors++
					}
				}

				if errors > 100 {
					t.Error("too many errors")
					return
				}
			}
		})
	}
}
