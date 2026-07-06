// Issue 7419: odd behavior for float constants underflowing to 0.
// Go silently rounds such constants to 0 when converted to float64.

package main

func main() {
	var x = 1e-779137
	var y = 1e-779138
	if x != 0 {
		panic("x != 0")
	}
	if y != 0 {
		panic("y != 0")
	}

	// Boundary: 1e-100001 is the first exponent past apd's range (rounds to 0
	// at parse); 1e-99999 still parses exactly and rounds to 0 on conversion.
	var a = 1e-100001
	var b = 1e-99999
	println(a == 0, b == 0)

	// Fractional digits count toward the effective exponent:
	// 0.001e-99999 is 1e-100002, past apd's range.
	var f = 0.001e-99999
	println(f == 0)

	// Hex-float literals underflow to 0 the same way.
	var h = 0x1p-999999
	println(h == 0)

	// Known divergence: the untyped constant rounds to 0 at parse time
	// because apd cannot represent exponents below -100000, so this prints
	// true; Go compares untyped constants exactly and prints false.
	const c = 1e-779137
	println(c == 0)

	println("ok")
}

// Output:
// true true
// true
// true
// true
// ok
