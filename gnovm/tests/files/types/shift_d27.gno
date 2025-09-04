package main

func main() {
	x := 11
	var y float32
	y = -(1 << x)
	println(y)
}

// Error:
// main/shift_d27.gno:6:2-15: operator << not defined on: Float32Kind

// TypeCheckError:
// main/shift_d27.gno:6:8: invalid operation: shifted operand 1 (type float32) must be integer
