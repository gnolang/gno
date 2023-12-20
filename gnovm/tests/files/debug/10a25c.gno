package main

func main() {
	x := 11
	var y float32
	y = -(1 << x)
	println(y)
}

// Error:
// main/files/debug/10a25c.gno:6: float not compatible with shift
