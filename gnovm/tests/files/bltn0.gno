package main

func main() {
	f := println
	f("Hello")
}

// Error:
// main/bltn0.gno:4:7-14: use of builtin println not in function call

// TypeCheckError:
// main/bltn0.gno:4:7: println (built-in) must be called
