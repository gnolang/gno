package main

func main() {
	r := "hello" << 1
	println(r)
}

// Error:
// main/files/types/shift_a16.gno:4:7: operator << not defined on: StringKind

// TypeCheckError:
// main/files/types/shift_a16.gno:4:7: invalid operation: shifted operand "hello" (untyped string constant) must be integer
