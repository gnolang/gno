package main

func main() {
	var s uint = 33
	var u2 = 1<<s + 1.0
	println(u2)
}
// Error:
// main/shift_f4c.gno:5:6-21: operator << not defined on: BigdecKind

// TypeCheckError:
// main/shift_f4c.gno:5:11: invalid operation: shifted operand 1 (type float64) must be integer