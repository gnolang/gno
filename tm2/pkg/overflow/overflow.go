// Package overflow offers overflow-checked arithmetic operations
// for all signed and unsigned integer types, as well as floating-point types.
// Each of the operations returns a result,bool combination.
//
// The functions support all types convertible to unsigned or signed integer
// types, and float32/float64 types. The modulo % operation is not present,
// as it is always safe.
package overflow

import "math"

// Number is a type constraint for all integer and floating-point values.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// isValidFloat checks if a float value is finite (not Inf or NaN).
// This is used to detect float overflow, which behaves differently from integer overflow:
// integers wrap around on overflow, while floats become ±Inf or NaN.
func isValidFloat[N Number](n N) bool {
	switch v := any(n).(type) {
	case float32:
		return !math.IsInf(float64(v), 0) && !math.IsNaN(float64(v))
	case float64:
		return !math.IsInf(v, 0) && !math.IsNaN(v)
	default:
		// For integers, always return true as they use different overflow detection
		return true
	}
}

// Add sums two numbers, returning the result and a boolean status.
// For integers, detects wrap-around overflow.
// For floats, detects when the result becomes Inf or NaN.
func Add[N Number](a, b N) (N, bool) {
	c := a + b
	switch any(c).(type) {
	case float32, float64:
		return c, isValidFloat(c)
	default:
		return c, (c > a) == (b > 0)
	}
}

// Sub returns the difference of two numbers and a boolean status.
// For integers, detects wrap-around overflow.
// For floats, detects when the result becomes Inf or NaN.
func Sub[N Number](a, b N) (N, bool) {
	c := a - b
	switch any(c).(type) {
	case float32, float64:
		return c, isValidFloat(c)
	default:
		return c, (c < a) == (b > 0)
	}
}

// Mul returns the multiplication of two numbers and a boolean status.
// For integers, detects wrap-around overflow.
// For floats, detects when the result becomes Inf or NaN.
func Mul[N Number](a, b N) (N, bool) {
	c := a * b
	switch any(c).(type) {
	case float32, float64:
		return c, isValidFloat(c)
	default:
		if a == 0 || b == 0 {
			return 0, true
		}
		return c, (c < 0) == ((a < 0) != (b < 0)) && (c/b == a)
	}
}

// Div returns the quotient of two numbers and a boolean status.
// For integers, detects wrap-around overflow.
// For floats, detects when the result becomes Inf or NaN.
func Div[N Number](a, b N) (N, bool) {
	if b == 0 {
		return 0, false
	}
	c := a / b
	switch any(c).(type) {
	case float32, float64:
		return c, isValidFloat(c)
	default:
		return c, c != a || b == 1 || a == 0
	}
}

// Addp returns the sum of two numbers, panicking on overflow.
func Addp[N Number](a, b N) N {
	r, ok := Add(a, b)
	if !ok {
		panic("addition overflow")
	}
	return r
}

// Subp returns the difference of two numbers, panicking on overflow.
func Subp[N Number](a, b N) N {
	r, ok := Sub(a, b)
	if !ok {
		panic("subtraction overflow")
	}
	return r
}

// Mulp returns the product of two numbers, panicking on overflow.
func Mulp[N Number](a, b N) N {
	r, ok := Mul(a, b)
	if !ok {
		panic("multiplication overflow")
	}
	return r
}

// Divp returns the quotient of two numbers, panicking on overflow.
func Divp[N Number](a, b N) N {
	r, ok := Div(a, b)
	if !ok {
		panic("division failure")
	}
	return r
}
