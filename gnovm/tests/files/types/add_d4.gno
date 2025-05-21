package main

var a int
var b any

func main() {
	println(b + a)
}

// Error:
// main/files/types/add_d4.gno:7:10: operator + not defined on: InterfaceKind

// TypeCheckError:
// main/files/types/add_d4.gno:7:10: invalid operation: b + a (mismatched types any and int)
