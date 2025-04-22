package main

func main() {
	println("hello")
	fallthrough
	println("world")
}

// Error:
// main/files/switch8.gno:5:2: fallthrough statement out of place

// TypeCheckError:
// main/files/switch8.gno:5:2: fallthrough statement out of place
