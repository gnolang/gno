package main

func main() {
	var a, b, c int16 = -1<<15, -1, 0
	c = a / b	// overflow: -32768 instead of 32768
	println(c)
}

// Error:
// division by zero or overflow
