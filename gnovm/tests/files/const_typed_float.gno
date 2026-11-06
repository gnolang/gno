package main

// Typed float constants fold exactly and round once to their type; the
// values equal IEEE arithmetic on the operands.

const a float32 = 0.1
const b float64 = 0.1
const big float64 = 1e308

func main() {
	println(a*3 == 0.3, b*3 == 0.30000000000000004, a+a+a == 0.3, b+b+b == 0.3)
	println(big/1e300 == 1e8, -big == -1e308, 1/b == 10)
	println(float64(a) == 0.1, float64(a) == 0.10000000149011612, float32(b) == a)
}

// Output:
// true true true false
// true true true
// false true true
