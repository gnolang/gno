package main

type S struct {
	T S
}

func main() {
	var a, b S
}

// Error:
// main/files/recursive1b.gno:1: type define cycle with name: S
