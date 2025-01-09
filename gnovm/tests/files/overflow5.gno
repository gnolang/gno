package main

func main() {
	var a, b, c int = -5, 7, 0
	c = a % b	// 0 quotient triggers a false negative in gnolang/overflow
	println(c)
}

// Output:
// -5
