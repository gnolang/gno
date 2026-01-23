package main

func main() {
	var s uint = 33

	var u1 bool
	u1 = 1.0<<s != 0
	println(u1)
}

// Error:
// main/shift_f1b.gno:7:2-18: operator << not defined on: Float64Kind

// TypeCheckError:
// main/shift_f1b.gno:7:7: invalid operation: shifted operand 1.0 (type float64) must be integer
