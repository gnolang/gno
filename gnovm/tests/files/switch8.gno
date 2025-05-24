package main

func main() {
	println("hello")
	fallthrough
	println("world")
}

// Error:
// main/switch8.gno:5:2-13: fallthrough statement out of place

// TypeCheckError:
// main/switch8.gno:5:2: fallthrough statement out of place
