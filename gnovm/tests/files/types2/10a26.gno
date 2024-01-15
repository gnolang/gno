package main

func main() {
	x := 11
	var y float32
	y = 1 + 1<<x + 1
	println(y)
}

// Error:
// main/files/types2/10a26.gno:6: operator << not defined on: Float32Kind
