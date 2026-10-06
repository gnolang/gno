package main

func main() {
	var s uint = 33
	var u2 = 1<<s != 1.0
	println(u2)
}

// Error:
// main/shift_f2.gno:5:11-22: operator << not defined on: Float64Kind

// TypeCheckError:
// main/shift_f2.gno:5:11: invalid operation: shifted operand 1 (type float64) must be integer
