package main

func main() {
	a := 1.0
	a <<= 1
	println(a)
}

// Error:
// main/shift_d51.gno:5:2-9: operator <<= not defined on: Float64Kind

// TypeCheckError:
// main/shift_d51.gno:5:2: invalid operation: shifted operand a (variable of type float64) must be integer
