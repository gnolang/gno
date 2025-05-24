package main

// TODO: fix in assignment
func main() {
	a := "hello"
	a <<= 1
}

// Error:
// main/shift_a15.gno:6:2-9: operator <<= not defined on: StringKind

// TypeCheckError:
// main/shift_a15.gno:6:2: invalid operation: shifted operand a (variable of type string) must be integer
