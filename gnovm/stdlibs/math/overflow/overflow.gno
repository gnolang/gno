// This is modified from https://github.com/JohnCGriffin/overflow (MIT).
// NOTE: there was a bug with the original Quotient* functions, and
// testing method. These have been fixed here, and tests ported to
// tests/files/maths_int*.go respectively.
// Note: moved over from p/demo/maths.

/*
Package overflow offers overflow-checked integer arithmetic operations
for int, int32, and int64. Each of the operations returns a
result,bool combination.  This was prompted by the need to know when
to flow into higher precision types from the math.big library.

For instance, assuing a 64 bit machine:

10 + 20 -> 30
int(math.MaxInt64) + 1 -> -9223372036854775808

whereas

overflow.Add(10,20) -> (30, true)
overflow.Add(math.MaxInt64,1) -> (0, false)

Add, Sub, Mul, Div are for int.  Add64, Add32, etc. are specifically sized.

If anybody wishes an unsigned version, submit a pull request for code
and new tests.
*/
package overflow

import "math"

//go:generate ./overflow_template.sh

func _is64Bit() bool {
	maxU32 := uint(math.MaxUint32)
	return ((maxU32 << 1) >> 1) == maxU32
}

/********** PARTIAL TEST COVERAGE FROM HERE DOWN *************

The only way that I could see to do this is a combination of
my normal 64 bit system and a GopherJS running on Node.  My
understanding is that its ints are 32 bit.

So, FEEL FREE to carefully review the code visually.

*************************************************************/

// Unspecified size, i.e. normal signed int

// Add sums two ints, returning the result and a boolean status.
func Add(a, b int) (int, bool) {
	if _is64Bit() {
		r64, ok := Add64(int64(a), int64(b))
		return int(r64), ok
	}
	r32, ok := Add32(int32(a), int32(b))
	return int(r32), ok
}

// Sub returns the difference of two ints and a boolean status.
func Sub(a, b int) (int, bool) {
	if _is64Bit() {
		r64, ok := Sub64(int64(a), int64(b))
		return int(r64), ok
	}
	r32, ok := Sub32(int32(a), int32(b))
	return int(r32), ok
}

// Mul returns the product of two ints and a boolean status.
func Mul(a, b int) (int, bool) {
	if _is64Bit() {
		r64, ok := Mul64(int64(a), int64(b))
		return int(r64), ok
	}
	r32, ok := Mul32(int32(a), int32(b))
	return int(r32), ok
}

// Div returns the quotient of two ints and a boolean status
func Div(a, b int) (int, bool) {
	if _is64Bit() {
		r64, ok := Div64(int64(a), int64(b))
		return int(r64), ok
	}
	r32, ok := Div32(int32(a), int32(b))
	return int(r32), ok
}

// Quo returns the quotient, remainder and status of two ints
func Quo(a, b int) (int, int, bool) {
	if _is64Bit() {
		q64, r64, ok := Quo64(int64(a), int64(b))
		return int(q64), int(r64), ok
	}
	q32, r32, ok := Quo32(int32(a), int32(b))
	return int(q32), int(r32), ok
}

/************* Panic versions for int ****************/

// Addp returns the sum of two ints, panicking on overflow
func Addp(a, b int) int {
	r, ok := Add(a, b)
	if !ok {
		panic("addition overflow")
	}
	return r
}

// Subp returns the difference of two ints, panicking on overflow.
func Subp(a, b int) int {
	r, ok := Sub(a, b)
	if !ok {
		panic("subtraction overflow")
	}
	return r
}

// Mulp returns the product of two ints, panicking on overflow.
func Mulp(a, b int) int {
	r, ok := Mul(a, b)
	if !ok {
		panic("multiplication overflow")
	}
	return r
}

// Divp returns the quotient of two ints, panicking on overflow.
func Divp(a, b int) int {
	r, ok := Div(a, b)
	if !ok {
		panic("division failure")
	}
	return r
}

//----------------------------------------
// This is generated code, created by overflow_template.sh executed
// by "go generate"

// Add8 performs + operation on two int8 operands
// returning a result and status
func Add8(a, b int8) (int8, bool) {
	c := a + b
	if (c > a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Add8p is the unchecked panicking version of Add8
func Add8p(a, b int8) int8 {
	r, ok := Add8(a, b)
	if !ok {
		panic("addition overflow")
	}
	return r
}

// Sub8 performs - operation on two int8 operands
// returning a result and status
func Sub8(a, b int8) (int8, bool) {
	c := a - b
	if (c < a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Sub8p is the unchecked panicking version of Sub8
func Sub8p(a, b int8) int8 {
	r, ok := Sub8(a, b)
	if !ok {
		panic("subtraction overflow")
	}
	return r
}

// Mul8 performs * operation on two int8 operands
// returning a result and status
func Mul8(a, b int8) (int8, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if (c < 0) == ((a < 0) != (b < 0)) {
		if c/b == a {
			return c, true
		}
	}
	return c, false
}

// Mul8p is the unchecked panicking version of Mul8
func Mul8p(a, b int8) int8 {
	r, ok := Mul8(a, b)
	if !ok {
		panic("multiplication overflow")
	}
	return r
}

// Div8 performs / operation on two int8 operands
// returning a result and status
func Div8(a, b int8) (int8, bool) {
	q, _, ok := Quo8(a, b)
	return q, ok
}

// Div8p is the unchecked panicking version of Div8
func Div8p(a, b int8) int8 {
	r, ok := Div8(a, b)
	if !ok {
		panic("division failure")
	}
	return r
}

// Quo8 performs + operation on two int8 operands
// returning a quotient, a remainder and status
func Quo8(a, b int8) (int8, int8, bool) {
	if b == 0 {
		return 0, 0, false
	} else if b == -1 && a == math.MinInt8 {
		return 0, 0, false
	}
	c := a / b
	return c, a % b, true
}

// Add16 performs + operation on two int16 operands
// returning a result and status
func Add16(a, b int16) (int16, bool) {
	c := a + b
	if (c > a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Add16p is the unchecked panicking version of Add16
func Add16p(a, b int16) int16 {
	r, ok := Add16(a, b)
	if !ok {
		panic("addition overflow")
	}
	return r
}

// Sub16 performs - operation on two int16 operands
// returning a result and status
func Sub16(a, b int16) (int16, bool) {
	c := a - b
	if (c < a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Sub16p is the unchecked panicking version of Sub16
func Sub16p(a, b int16) int16 {
	r, ok := Sub16(a, b)
	if !ok {
		panic("subtraction overflow")
	}
	return r
}

// Mul16 performs * operation on two int16 operands
// returning a result and status
func Mul16(a, b int16) (int16, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if (c < 0) == ((a < 0) != (b < 0)) {
		if c/b == a {
			return c, true
		}
	}
	return c, false
}

// Mul16p is the unchecked panicking version of Mul16
func Mul16p(a, b int16) int16 {
	r, ok := Mul16(a, b)
	if !ok {
		panic("multiplication overflow")
	}
	return r
}

// Div16 performs / operation on two int16 operands
// returning a result and status
func Div16(a, b int16) (int16, bool) {
	q, _, ok := Quo16(a, b)
	return q, ok
}

// Div16p is the unchecked panicking version of Div16
func Div16p(a, b int16) int16 {
	r, ok := Div16(a, b)
	if !ok {
		panic("division failure")
	}
	return r
}

// Quo16 performs + operation on two int16 operands
// returning a quotient, a remainder and status
func Quo16(a, b int16) (int16, int16, bool) {
	if b == 0 {
		return 0, 0, false
	} else if b == -1 && a == math.MinInt16 {
		return 0, 0, false
	}
	c := a / b
	return c, a % b, true
}

// Add32 performs + operation on two int32 operands
// returning a result and status
func Add32(a, b int32) (int32, bool) {
	c := a + b
	if (c > a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Add32p is the unchecked panicking version of Add32
func Add32p(a, b int32) int32 {
	r, ok := Add32(a, b)
	if !ok {
		panic("addition overflow")
	}
	return r
}

// Sub32 performs - operation on two int32 operands
// returning a result and status
func Sub32(a, b int32) (int32, bool) {
	c := a - b
	if (c < a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Sub32p is the unchecked panicking version of Sub32
func Sub32p(a, b int32) int32 {
	r, ok := Sub32(a, b)
	if !ok {
		panic("subtraction overflow")
	}
	return r
}

// Mul32 performs * operation on two int32 operands
// returning a result and status
func Mul32(a, b int32) (int32, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if (c < 0) == ((a < 0) != (b < 0)) {
		if c/b == a {
			return c, true
		}
	}
	return c, false
}

// Mul32p is the unchecked panicking version of Mul32
func Mul32p(a, b int32) int32 {
	r, ok := Mul32(a, b)
	if !ok {
		panic("multiplication overflow")
	}
	return r
}

// Div32 performs / operation on two int32 operands
// returning a result and status
func Div32(a, b int32) (int32, bool) {
	q, _, ok := Quo32(a, b)
	return q, ok
}

// Div32p is the unchecked panicking version of Div32
func Div32p(a, b int32) int32 {
	r, ok := Div32(a, b)
	if !ok {
		panic("division failure")
	}
	return r
}

// Quo32 performs + operation on two int32 operands
// returning a quotient, a remainder and status
func Quo32(a, b int32) (int32, int32, bool) {
	if b == 0 {
		return 0, 0, false
	} else if b == -1 && a == math.MinInt32 {
		return 0, 0, false
	}
	c := a / b
	return c, a % b, true
}

// Add64 performs + operation on two int64 operands
// returning a result and status
func Add64(a, b int64) (int64, bool) {
	c := a + b
	if (c > a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Add64p is the unchecked panicking version of Add64
func Add64p(a, b int64) int64 {
	r, ok := Add64(a, b)
	if !ok {
		panic("addition overflow")
	}
	return r
}

// Sub64 performs - operation on two int64 operands
// returning a result and status
func Sub64(a, b int64) (int64, bool) {
	c := a - b
	if (c < a) == (b > 0) {
		return c, true
	}
	return c, false
}

// Sub64p is the unchecked panicking version of Sub64
func Sub64p(a, b int64) int64 {
	r, ok := Sub64(a, b)
	if !ok {
		panic("subtraction overflow")
	}
	return r
}

// Mul64 performs * operation on two int64 operands
// returning a result and status
func Mul64(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if (c < 0) == ((a < 0) != (b < 0)) {
		if c/b == a {
			return c, true
		}
	}
	return c, false
}

// Mul64p is the unchecked panicking version of Mul64
func Mul64p(a, b int64) int64 {
	r, ok := Mul64(a, b)
	if !ok {
		panic("multiplication overflow")
	}
	return r
}

// Div64 performs / operation on two int64 operands
// returning a result and status
func Div64(a, b int64) (int64, bool) {
	q, _, ok := Quo64(a, b)
	return q, ok
}

// Div64p is the unchecked panicking version of Div64
func Div64p(a, b int64) int64 {
	r, ok := Div64(a, b)
	if !ok {
		panic("division failure")
	}
	return r
}

// Quo64 performs + operation on two int64 operands
// returning a quotient, a remainder and status
func Quo64(a, b int64) (int64, int64, bool) {
	if b == 0 {
		return 0, 0, false
	} else if b == -1 && a == math.MinInt64 {
		return 0, 0, false
	}
	c := a / b
	return c, a % b, true
}
