package main

func main() {
	i := int(1)
	a := +(1.0 << i)

	println(a)
}

// Error:
// main/shift_k3.gno:5:2-18: operator << not defined on: Float64Kind

// TypeCheckError:
// main/shift_k3.gno:5:9: invalid operation: shifted operand 1.0 (type float64) must be integer
