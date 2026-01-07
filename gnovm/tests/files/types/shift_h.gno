package main

func main() {
	x := 2
	y := 1.0<<x + 1
	println(y)
}

// Error:
// main/shift_h.gno:5:2-17: invalid operation: shifted operand (const (1.0 <untyped> bigdec)) (<untyped> bigdec) must be integer

// TypeCheckError:
// main/shift_h.gno:5:7: invalid operation: shifted operand 1.0 (type float64) must be integer
