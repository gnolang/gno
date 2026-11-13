package main

// An untyped integer constant converts to float32 rounded once, from the
// exact value, ties to even (#6394): 1<<54 + 1<<30 is a float32 midpoint,
// one more lies above it and rounds up.

const F = float32(1<<54 + 1<<30 + 1)

func main() {
	println(int64(float32(1<<54 + 1<<30 + 1)))
	println(int64(float32(-(1<<54 + 1<<30 + 1))))
	println(int64(F), int64(float64(F)))
	println(int64(float32(1<<54+1<<30)), int64(float32(1<<54+1<<30-1)))
	println(int64(float32(1<<24+1)), int64(float32(1<<24+3)))
	var f32 float32 = 1<<54 + 1<<30 + 1
	var f64 float64 = 1<<54 + 1<<30 + 1
	println(int64(f32), int64(f64))
	println(float32(1<<54+1<<30+1) == 1<<54+2<<30)
}

// Output:
// 18014400656965632
// -18014400656965632
// 18014400656965632 18014400656965632
// 18014398509481984 18014398509481984
// 16777216 16777220
// 18014400656965632 18014399583223808
// true
