// This is modified from https://github.com/JohnCGriffin/overflow (MIT).
// Updated from https://github.com/rwxe/overflow (MIT).
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

// Abs get absolute value of an int, returning the result and a ok result indicating whether the operation is safe.
func Abs(x int) (int, bool) {
	if _is64Bit() {
		r64, ok := Abs64(int64(x))
		return int(r64), ok
	}
	r32, ok := Abs32(int32(x))
	return int(r32), ok
}

// Add sums two ints, returning the result and a boolean status.
func Add(a, b int) (int, bool) {
	if _is64Bit() {
		r64, ok := Add64(int64(a), int64(b))
		return int(r64), ok
	}
	r32, ok := Add32(int32(a), int32(b))
	return int(r32), ok
}

// UAdd sums two uints, returning the result and a ok result indicating whether the operation is safe.
func UAdd(a, b uint) (uint, bool) {
	if _is64Bit() {
		r64, ok := UAdd64(uint64(a), uint64(b))
		return uint(r64), ok
	}
	r32, ok := UAdd32(uint32(a), uint32(b))
	return uint(r32), ok
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

// USub returns the difference of two uints and a ok result indicating whether the operation is safe.
func USub(a, b uint) (uint, bool) {
	if _is64Bit() {
		r64, ok := USub64(uint64(a), uint64(b))
		return uint(r64), ok
	}
	r32, ok := USub32(uint32(a), uint32(b))
	return uint(r32), ok
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

// UMul returns the product of two uints and a ok result indicating whether the operation is safe.
func UMul(a, b uint) (uint, bool) {
	if _is64Bit() {
		r64, ok := UMul64(uint64(a), uint64(b))
		return uint(r64), ok
	}
	r32, ok := UMul32(uint32(a), uint32(b))
	return uint(r32), ok
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

// UDiv returns the quotient of two uints and a ok result indicating whether the operation is safe.
func UDiv(a, b uint) (uint, bool) {
	if _is64Bit() {
		r64, ok := UDiv64(uint64(a), uint64(b))
		return uint(r64), ok
	}
	r32, ok := UDiv32(uint32(a), uint32(b))
	return uint(r32), ok
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

// UQuo returns the quotient, remainder and ok result indicating whether the operation is safe.
func UQuo(a, b uint) (uint, uint, bool) {
	if _is64Bit() {
		uq64, ur64, ok := UQuo64(uint64(a), uint64(b))
		return uint(uq64), uint(ur64), ok
	}
	uq32, ur32, ok := UQuo32(uint32(a), uint32(b))
	return uint(uq32), uint(ur32), ok
}

/************* Panic versions for int ****************/

// Absp returns the absolute value, panicking on overflow
func Absp(x int) int {
	r, ok := Abs(x)
	if !ok {
		panic("absolute value overflow")
	}
	return r
}

// Addp returns the sum of two ints, panicking on overflow
func Addp(a, b int) int {
	r, ok := Add(a, b)
	if !ok {
		panic("addition overflow")
	}
	return r
}

// UAddp returns the sum of two uints, panicking on overflow
func UAddp(a, b uint) uint {
	r, ok := UAdd(a, b)
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

// USubp returns the difference of two uints, panicking on overflow.
func USubp(a, b uint) uint {
	r, ok := USub(a, b)
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

// UMulp returns the product of two uints, panicking on overflow.
func UMulp(a, b uint) uint {
	r, ok := UMul(a, b)
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

// UDivp returns the quotient of two uints, panicking on overflow.
func UDivp(a, b uint) uint {
	r, ok := UDiv(a, b)
	if !ok {
		panic("division failure")
	}
	return r
}

//----------------------------------------
// This is generated code, created by overflow_template.sh executed
// by "go generate"

// Abs8 performs absolute value operation on an int8 operand
// returning a result and a ok result indicating whether the operation is safe.
func Abs8(x int8) (int8, bool) {
	if x == math.MinInt8 {
		return x, false
	}
	if x >= 0 {
		return x, true
	}
	return -x, true
}

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

// UAdd8 performs + operation on two uint8 operands
// returning a result and a ok result indicating whether the operation is safe.
func UAdd8(a, b uint8) (uint8, bool) {
	c := a + b
	if c >= a {
		return c, true
	}
	return c, false
}

// UAdd8p is the unchecked panicing version of UAdd8
func UAdd8p(a, b uint8) uint8 {
	r, ok := UAdd8(a, b)
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

// USub8 performs - operation on two uint8 operands
// returning a result and a ok result indicating whether the operation is safe.
func USub8(a, b uint8) (uint8, bool) {
	c := a - b
	if a >= b {
		return c, true
	}
	return c, false
}

// USub8p is the unchecked panicing version of USub8
func USub8p(a, b uint8) uint8 {
	r, ok := USub8(a, b)
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

// UMul8 performs * operation on two uint8 operands
// returning a result and a ok result indicating whether the operation is safe.
func UMul8(a, b uint8) (uint8, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if c/b == a {
		return c, true
	}
	return c, false
}

// UMul8p is the unchecked panicing version of UMul8
func UMul8p(a, b uint8) uint8 {
	r, ok := UMul8(a, b)
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

// UDiv8 performs / operation on two uint8 operands
// returning a result and a ok result indicating whether the operation is safe.
func UDiv8(a, b uint8) (uint8, bool) {
	q, _, ok := UQuo8(a, b)
	return q, ok
}

// UDiv8p is the unchecked panicing version of UDiv8
func UDiv8p(a, b uint8) uint8 {
	r, ok := UDiv8(a, b)
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

// UQuo8 performs + operation on two uint8 operands
// returning a quotient, a remainder and a ok result indicating whether the operation is safe.
func UQuo8(a, b uint8) (uint8, uint8, bool) {
	if b == 0 {
		return 0, 0, false
	}
	c := a / b
	return c, a % b, true
}

// Abs16 performs absolute value operation on an int16 operand
// returning a result and a ok result indicating whether the operation is safe.
func Abs16(x int16) (int16, bool) {
	if x == math.MinInt16 {
		return x, false
	}
	if x >= 0 {
		return x, true
	}
	return -x, true
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

// UAdd16 performs + operation on two uint16 operands
// returning a result and a ok result indicating whether the operation is safe.
func UAdd16(a, b uint16) (uint16, bool) {
	c := a + b
	if c >= a {
		return c, true
	}
	return c, false
}

// UAdd16p is the unchecked panicing version of UAdd16
func UAdd16p(a, b uint16) uint16 {
	r, ok := UAdd16(a, b)
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

// USub16 performs - operation on two uint16 operands
// returning a result and a ok result indicating whether the operation is safe.
func USub16(a, b uint16) (uint16, bool) {
	c := a - b
	if a >= b {
		return c, true
	}
	return c, false
}

// USub16p is the unchecked panicing version of USub16
func USub16p(a, b uint16) uint16 {
	r, ok := USub16(a, b)
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

// UMul16 performs * operation on two uint16 operands
// returning a result and a ok result indicating whether the operation is safe.
func UMul16(a, b uint16) (uint16, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if c/b == a {
		return c, true
	}
	return c, false
}

// UMul16p is the unchecked panicing version of UMul16
func UMul16p(a, b uint16) uint16 {
	r, ok := UMul16(a, b)
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

// UDiv16 performs / operation on two uint16 operands
// returning a result and a ok result indicating whether the operation is safe.
func UDiv16(a, b uint16) (uint16, bool) {
	q, _, ok := UQuo16(a, b)
	return q, ok
}

// UDiv16p is the unchecked panicing version of UDiv16
func UDiv16p(a, b uint16) uint16 {
	r, ok := UDiv16(a, b)
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

// UQuo16 performs + operation on two uint16 operands
// returning a quotient, a remainder and a ok result indicating whether the operation is safe.
func UQuo16(a, b uint16) (uint16, uint16, bool) {
	if b == 0 {
		return 0, 0, false
	}
	c := a / b
	return c, a % b, true
}

// Abs32 performs absolute value operation on an int32 operand
// returning a result and a ok result indicating whether the operation is safe.
func Abs32(x int32) (int32, bool) {
	if x == math.MinInt32 {
		return x, false
	}
	if x >= 0 {
		return x, true
	}
	return -x, true
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

// UAdd32 performs + operation on two uint32 operands
// returning a result and a ok result indicating whether the operation is safe.
func UAdd32(a, b uint32) (uint32, bool) {
	c := a + b
	if c >= a {
		return c, true
	}
	return c, false
}

// UAdd32p is the unchecked panicing version of UAdd32
func UAdd32p(a, b uint32) uint32 {
	r, ok := UAdd32(a, b)
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

// USub32 performs - operation on two uint32 operands
// returning a result and a ok result indicating whether the operation is safe.
func USub32(a, b uint32) (uint32, bool) {
	c := a - b
	if a >= b {
		return c, true
	}
	return c, false
}

// USub32p is the unchecked panicing version of USub32
func USub32p(a, b uint32) uint32 {
	r, ok := USub32(a, b)
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

// UMul32 performs * operation on two uint32 operands
// returning a result and a ok result indicating whether the operation is safe.
func UMul32(a, b uint32) (uint32, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if c/b == a {
		return c, true
	}
	return c, false
}

// UMul32p is the unchecked panicing version of UMul32
func UMul32p(a, b uint32) uint32 {
	r, ok := UMul32(a, b)
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

// UDiv32 performs / operation on two uint32 operands
// returning a result and a ok result indicating whether the operation is safe.
func UDiv32(a, b uint32) (uint32, bool) {
	q, _, ok := UQuo32(a, b)
	return q, ok
}

// UDiv32p is the unchecked panicing version of UDiv32
func UDiv32p(a, b uint32) uint32 {
	r, ok := UDiv32(a, b)
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

// UQuo32 performs + operation on two uint32 operands
// returning a quotient, a remainder and a ok result indicating whether the operation is safe.
func UQuo32(a, b uint32) (uint32, uint32, bool) {
	if b == 0 {
		return 0, 0, false
	}
	c := a / b
	return c, a % b, true
}

// Abs64 performs absolute value operation on an int64 operand
// returning a result and a ok result indicating whether the operation is safe.
func Abs64(x int64) (int64, bool) {
	if x == math.MinInt64 {
		return x, false
	}
	if x >= 0 {
		return x, true
	}
	return -x, true
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

// UAdd64 performs + operation on two uint64 operands
// returning a result and a ok result indicating whether the operation is safe.
func UAdd64(a, b uint64) (uint64, bool) {
	c := a + b
	if c >= a {
		return c, true
	}
	return c, false
}

// UAdd64p is the unchecked panicing version of UAdd64
func UAdd64p(a, b uint64) uint64 {
	r, ok := UAdd64(a, b)
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

// USub64 performs - operation on two uint64 operands
// returning a result and a ok result indicating whether the operation is safe.
func USub64(a, b uint64) (uint64, bool) {
	c := a - b
	if a >= b {
		return c, true
	}
	return c, false
}

// USub64p is the unchecked panicing version of USub64
func USub64p(a, b uint64) uint64 {
	r, ok := USub64(a, b)
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

// UMul64 performs * operation on two uint64 operands
// returning a result and a ok result indicating whether the operation is safe.
func UMul64(a, b uint64) (uint64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	if c/b == a {
		return c, true
	}
	return c, false
}

// UMul64p is the unchecked panicing version of UMul64
func UMul64p(a, b uint64) uint64 {
	r, ok := UMul64(a, b)
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

// UDiv64 performs / operation on two uint64 operands
// returning a result and a ok result indicating whether the operation is safe.
func UDiv64(a, b uint64) (uint64, bool) {
	q, _, ok := UQuo64(a, b)
	return q, ok
}

// UDiv64p is the unchecked panicing version of UDiv64
func UDiv64p(a, b uint64) uint64 {
	r, ok := UDiv64(a, b)
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

// UQuo64 performs + operation on two uint64 operands
// returning a quotient, a remainder and a ok result indicating whether the operation is safe.
func UQuo64(a, b uint64) (uint64, uint64, bool) {
	if b == 0 {
		return 0, 0, false
	}
	c := a / b
	return c, a % b, true
}
