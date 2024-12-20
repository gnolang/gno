package main

func main() {
	var a, b, c int32 = -1<<31, -1, 0
	c = a / b	// overflow: -2147483648 instead of 2147483648
	println(c)
}

// Error:
// division by zero or overflow
