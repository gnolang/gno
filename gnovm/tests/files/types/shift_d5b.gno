package main

func main() {
	x := 11
	y := float32(1.0 << x)
	println(y)
}

// Error:
// main/files/types/shift_d5b.gno:5:7: operator << not defined on: Float32Kind
