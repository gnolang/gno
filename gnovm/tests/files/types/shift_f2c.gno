package main

func main() {
	var s uint = 33
	var u2 = 1 != 1.0<<s
	println(u2)
}

// Error:
// main/shift_f2c.gno:5:11-22: operator << not defined on: Float64Kind

// TypeCheckError:
// main/shift_f2c.gno:5:16: invalid operation: shifted operand 1.0 (type float64) must be integer
