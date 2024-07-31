package main

func main() {
	x := 11
	var y float32
	y = 1 + 1<<x + 1
	println(y)
}

// Error:
// main/files/types2/shift_d28.gno:6:2: operator << not defined on: Float32Kind
