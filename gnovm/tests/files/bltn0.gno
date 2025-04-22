package main

func main() {
	f := println
	f("Hello")
}

// Error:
// main/files/bltn0.gno:4:7: use of builtin println not in function call

// TypeCheckError:
// main/files/bltn0.gno:4:7: println (built-in) must be called
