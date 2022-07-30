package main

func Bar() string {
	return
}

func main() {
	println(Bar())
}

// Error:
// main/files2/fun21.gno:4: expected 1 return values; got 0