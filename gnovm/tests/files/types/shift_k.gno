package main

func main() {
	var a uint = 1

	b := make([]byte, 1.0<<a)

	println(b)
}

// Error:
// main/shift_k.gno:6:7-27: operator << not defined on: Float64Kind
