package main

func main() {
	x := 11
	var y float32
	y = 1 + 1<<x + 1
	println(y)
}

// Error:
// main/shift_d28.gno:6:2-18: operator << not defined on: Float32Kind

// TypeCheckError:
// main/shift_d28.gno:6:10: invalid operation: shifted operand 1 (type float32) must be integer
