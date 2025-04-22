package main

func main() {
	var s uint = 33
	var u2 = 1.0 + 1<<s
	println(u2)
}

// Error:
// main/files/types/shift_f3b.gno:5:6: operator << not defined on: BigdecKind

// TypeCheckError:
// main/files/types/shift_f3b.gno:5:17: invalid operation: shifted operand 1 (type float64) must be integer
