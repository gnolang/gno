package main

func main() {
	x := 11
	y := float32(1 + 1<<x)
	println(y)
}

// Error:
// main/files/types2/shift_d20.gno:5: operator << not defined on: Float32Kind
