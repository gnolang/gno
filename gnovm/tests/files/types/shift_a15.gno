package main

// TODO: fix in assignment
func main() {
	a := "hello"
	a <<= 1
}

// Error:
// main/files/types/shift_a15.gno:6:2: operator <<= not defined on: StringKind

// TypeCheckError:
// main/files/types/shift_a15.gno:6:2: invalid operation: shifted operand a (variable of type string) must be integer
