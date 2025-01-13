package main

func main() {
	var a, b, c int8 = -1<<7, -1, 0
	c = a / b	// overflow: -128 instead of 128
	println(c)
}

// Error:
// division by zero or overflow
