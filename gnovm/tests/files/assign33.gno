package main

func main() {
	n := f()
}

func f() {
	println("hello!")
}

// Error:
// main/gnovm/tests/files/assign33.gno:3:1: f() (no value) used as value
