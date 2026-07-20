package main

// Untyped float constants past the ratOverflowBits threshold (~1e±1233 in
// magnitude) must NOT panic: go/constant switches from big.Rat to a bounded
// big.Float in that range, and Gno matches that behavior via automatic
// promotion in wrapRatOrPromote/parseBigdecLiteral. This prevents both the
// unbounded-numerator DoS (repeated squaring of a huge integer constant)
// and the unbounded-denominator memory blowup (repeated squaring of 1/3).

func main() {
	// Cancellation across huge magnitudes: exponents fold, result is 1.
	println(1e10000 / 1e10000)

	// Direct assignment of a tiny constant to float64: underflows to 0.
	// (Signed-zero semantics for the negative variant are covered by float10.)
	const tiny = 1e-10000
	var f float64 = tiny
	println(f)

	// Repeated squaring of 1/3 stays within bounded memory: the rat form
	// overflows on the way up, but promotion to big.Float caps the storage.
	const a0 = 1.0 / 3.0
	const a1 = a0 * a0
	const a2 = a1 * a1
	const a3 = a2 * a2
	const a4 = a3 * a3
	const a5 = a4 * a4
	const a6 = a5 * a5
	const a7 = a6 * a6
	const a8 = a7 * a7
	const a9 = a8 * a8
	const a10 = a9 * a9
	const a11 = a10 * a10
	const a12 = a11 * a11
	// After promotion the value is effectively zero when narrowed to float64.
	var g float64 = a12
	println(g == 0)
}

// Output:
// 1
// 0
// true
