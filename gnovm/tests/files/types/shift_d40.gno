package main

func main() {
	a := 2
	s := []float32{1 << a}
	println(s[0])
}

// Error:
// main/shift_d40.gno:5:7-24: operator << not defined on: Float32Kind

// TypeCheckError:
// main/shift_d40.gno:5:17: invalid operation: shifted operand 1 (type float32) must be integer
