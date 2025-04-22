package main

func main() {
	var s uint = 33
	var u2 = 1 != 1.0<<s
	println(u2)
}

// Error:
// main/files/types/shift_f2c.gno:5:6: operator << not defined on: BigdecKind

// TypeCheckError:
// main/files/types/shift_f2c.gno:5:16: invalid operation: shifted operand 1.0 (type float64) must be integer
