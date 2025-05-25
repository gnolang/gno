package main

func main() {
	a := "hello"
	r := a << 1
	println(r)
}

// Error:
// main/shift_a14.gno:5:7-13: operator << not defined on: StringKind

// TypeCheckError:
// main/shift_a14.gno:5:7: invalid operation: shifted operand a (variable of type string) must be integer
