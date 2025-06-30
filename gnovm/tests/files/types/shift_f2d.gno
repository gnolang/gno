package main

func main() {
	var s uint = 33
	var u2 = 1<<s != 1.0<<s
	println(u2)
}

// Error:
// main/shift_f2d.gno:5:11-25: incompatible types in binary expression: <untyped> bigint NEQ <untyped> bigdec

// TypeCheckError:
// main/shift_f2d.gno:5:11: invalid operation: shifted operand 1 (type float64) must be integer; main/shift_f2d.gno:5:19: invalid operation: shifted operand 1.0 (type float64) must be integer
