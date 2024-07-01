package main

func main() {
	r := "hello" << 1
	println(r)
}

// Error:
// main/files/types2/shift_a16.gno:4: operator << not defined on: StringKind
