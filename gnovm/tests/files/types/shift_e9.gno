package main

func main() {
	x := -1
	y := 1
	y <<= x
	println(y)
}

// Error:
// runtime error: negative shift amount:(-1 int)
