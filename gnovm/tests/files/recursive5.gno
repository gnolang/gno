package main

type S struct {
	S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// main/files/recursive5.gno:1: type define cycle with name: S
