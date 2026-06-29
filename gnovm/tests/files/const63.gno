package main

// Untyped float constant arithmetic must use exact rational semantics,
// matching Go's math/big.Rat behavior. Each assertion below is verified
// to also hold in Go (run: go run _verify/const63_go/main.go).

func main() {
	// basic rational: 1/3 * 3 == 1 (fails with decimal arithmetic)
	const a = (1.0 / 3.0) * 3.0
	println(a == 1.0) // true

	// repeated thirds
	const b = 1.0/3.0 + 1.0/3.0 + 1.0/3.0
	println(b == 1.0) // true

	// sevenths
	const c = (1.0 / 7.0) * 7.0
	println(c == 1.0) // true

	// subtraction recovering zero
	const d = 1.0 - 1.0/3.0 - 1.0/3.0 - 1.0/3.0
	println(d == 0.0) // true

	// decimal literals that happen to be exact rationals
	const e = 0.1 + 0.2
	println(e == 0.3) // true
}

// Output:
// true
// true
// true
// true
// true
