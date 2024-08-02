package main

func main() {
	a := 1.0
	a <<= 1
	println(a)
}

// Error:
// main/files/types2/shift_d51.gno:5:2: operator <<= not defined on: Float64Kind
