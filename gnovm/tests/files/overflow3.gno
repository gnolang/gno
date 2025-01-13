package main

func main() {
	var a, b, c int64 = -1<<63, -1, 0
	c = a / b	// overflow: -9223372036854775808 instead of 9223372036854775808
	println(c)
}

// Error:
// division by zero or overflow
