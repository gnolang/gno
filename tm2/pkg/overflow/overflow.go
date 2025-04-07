// Package overflow offers overflow-checked integer arithmetic operations
// for all signed and unsigned integer types. Each of the operations returns a
// result,bool combination.
//
// The functions support all types convertible to unsigned or signed integer
// types. The modulo % operation is not present, as it is always safe.
package overflow

// Number is a type constraint for all integer values, signed and unsigned.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Add sums two numbers, returning the result and a boolean status.
func Add[N Number](a, b N) (N, bool) {
	c := a + b
	return c, (c > a) == (b > 0)
}

// Sub returns the difference of two numbers and a boolean status.
func Sub[N Number](a, b N) (N, bool) {
	c := a - b
	return c, (c < a) == (b > 0)
}

// Mul returns the multiplication of two numbers and a boolean status.
func Mul[N Number](a, b N) (N, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	c := a * b
	return c, (c < 0) == ((a < 0) != (b < 0)) && (c/b == a)
}

// Div returns the quotient of two numbers and a boolean status.
func Div[N Number](a, b N) (N, bool) {
	if b == 0 {
		return 0, false
	}
	// The only overflow case is 2^(bits-1)/-1, but we cannot use -1 with
	// generics which accept uints.
	// Thus, use another property: c == a can be the same for 0/N == 0 or
	// N/1 == N. The overflow operation above also results in c == a
	c := a / b
	return c, c != a || b == 1 || a == 0
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
