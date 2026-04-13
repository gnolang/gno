package main

func main() {
	r := "hello" << 1
	println(r)
}

// Error:
// main/shift_a16.gno:4:7-19: invalid operation: shifted operand (const ("hello" <untyped> string)) (<untyped> string) must be integer

// TypeCheckError:
// main/shift_a16.gno:4:7: invalid operation: shifted operand "hello" (untyped string constant) must be integer
