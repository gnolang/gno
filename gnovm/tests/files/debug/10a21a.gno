package main

func main() {
	x := 11
	y := float32(1 + 1<<x)
	println(y)
}

// Error:
// main/files/debug/10a21a.gno:5: float not compatible with shift
