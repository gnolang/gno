package main

func main() {
	var s uint = 33
	var u2 = 1.0 != 1<<s
	println(u2)
}

// Error:
// main/shift_f2b.gno:5:6-22: operator << not defined on: BigdecKind

// TypeCheckError:
// main/shift_f2b.gno:5:18: invalid operation: shifted operand 1 (type float64) must be integer
