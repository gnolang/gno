package main

func main() {
	x := 11
	y := float32(1.0 << x)
	println(y)
}

// Error:
// main/shift_d5b.gno:5:7-24: operator << not defined on: Float32Kind

// TypeCheckError:
// main/shift_d5b.gno:5:15: invalid operation: shifted operand 1.0 (type float32) must be integer
