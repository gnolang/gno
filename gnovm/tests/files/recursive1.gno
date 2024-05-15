package main

type S struct {
	T S
}

func main() {
	var a, b S
	println(a == b)
}

// Error:
// loop in variable initialization: dependency trail [S] circularly depends on S
