package main

func main() {
	var s uint = 32
	x := ('a' << s) + 1.5
	println(x)
}

// Error:
// main/shift_context7.gno:5:2-23: operator << not defined on: Float64Kind

// TypeCheckError:
// main/shift_context7.gno:5:8: invalid operation: shifted operand 'a' (type float64) must be integer
