// run

// Transcribed from Go's go/constant tests and issues #6389 #6392 #6393 #6394; expected output from go1.25.9.
// Converting an untyped integer constant to float32 rounds once, from the
// exact value, ties to even (#6394). 1<<54 + 1<<30 is a float32 midpoint;
// one more lies above it and must round up.

package main

const F = float32(1<<54 + 1<<30 + 1)

func main() {
	println(int64(float32(1<<54 + 1<<30 + 1)))
	println(int64(float32(-(1<<54 + 1<<30 + 1))))
	println(int64(F))
	println(int64(float64(float32(1<<54 + 1<<30 + 1))))

	// Midpoints: ties to even.
	println(int64(float32(1<<54 + 1<<30)))
	println(int64(float32(1<<54 + 1<<30 - 1)))
	println(int64(float32(1<<54 + 3<<30)))
	println(int64(float32(1<<24+1)), int64(float32(1<<24+3)))
	println(int64(float32(1<<25+2)), int64(float32(1<<25+6)))

	// float64 rounds once to 53 bits.
	println(int64(float64(1<<53+1)), int64(float64(1<<53+3)))
	println(int64(float64(1<<62+1<<9+1)), int64(float64(1<<62+1<<9)))
	println(float64(1<<54+1<<30+1) == 1<<54+1<<30)
	println(float32(1<<54+1<<30+1) == 1<<54+2<<30)
	println(float32(1<<54+1<<30) == 1<<54)

	// Typed variable declarations take the same path.
	var f32 float32 = 1<<54 + 1<<30 + 1
	var f64 float64 = 1<<54 + 1<<30 + 1
	println(int64(f32), int64(f64))
}

// GnoOutput:
// 18014398509481984
// -18014398509481984
// 18014398509481984
// 18014398509481984
// 18014398509481984
// 18014398509481984
// 18014402804449280
// 16777216 16777220
// 33554432 33554440
// 9007199254740992 9007199254740996
// 4611686018427388928 4611686018427387904
// true
// false
// true
// 18014398509481984 18014399583223808

// GoOutput:
// 18014400656965632
// -18014400656965632
// 18014400656965632
// 18014400656965632
// 18014398509481984
// 18014398509481984
// 18014402804449280
// 16777216 16777220
// 33554432 33554440
// 9007199254740992 9007199254740996
// 4611686018427388928 4611686018427387904
// true
// true
// true
// 18014400656965632 18014399583223808

// Verdict: KnownIssue run: #6394 untyped integer constant to float32 rounds twice (53 bits, then 24)
