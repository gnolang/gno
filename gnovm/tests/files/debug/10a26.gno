package main

func main() {
	x := 11
	var y float32
	y = 1 + 1<<x + 1
	println(y)
}

// Error:
// main/files/debug/10a26.gno:6: float not compatible with shift
