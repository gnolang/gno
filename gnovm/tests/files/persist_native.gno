package main

var a any

func init() {
	a = println
}

func main() {
	a("doesntwork")
}

// Error:
// main/persist_native.gno:6:6-13: use of builtin println not in function call

// TypeCheckError:
// main/persist_native.gno:6:6: println (built-in) must be called; main/persist_native.gno:10:2: invalid operation: cannot call a (variable of interface type any): any is not a function
