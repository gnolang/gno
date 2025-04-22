package main

func main() {
	a := "hello"
	r := a << 1
	println(r)
}

// Error:
// main/files/types/shift_a14.gno:5:7: operator << not defined on: StringKind

// TypeCheckError:
// main/files/types/shift_a14.gno:5:7: invalid operation: shifted operand a (variable of type string) must be integer
