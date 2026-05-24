package main

func main() {
	x := 11
	a := 1.0
	y := a << x 
	println(y)
}

// Error:
// main/shift_d5c.gno:6:7-13: operator << not defined on: Float64Kind

// TypeCheckError:
// main/shift_d5c.gno:6:7: invalid operation: shifted operand a (variable of type float64) must be integer
