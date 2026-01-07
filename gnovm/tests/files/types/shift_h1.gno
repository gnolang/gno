package main

func main() {
	x := 2
	y := 1 + 1.0<<x
	println(y)
}

// Error:
// main/shift_h1.gno:5:2-17: invalid operation: shifted operand (const (1.0 <untyped> bigdec)) (<untyped> bigdec) must be integer

// TypeCheckError:
// main/shift_h1.gno:5:11: invalid operation: shifted operand 1.0 (type float64) must be integer
