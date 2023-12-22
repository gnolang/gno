package main

func main() {
	x := 11
	var y float32
	y = -(1 << x)
	println(y)
}

// Error:
// operators << and <<= not defined for float32
