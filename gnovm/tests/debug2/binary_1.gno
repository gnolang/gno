package main

func main() {
	println(1 - "a")
}

// Error:
// main/debug/binary_1.gno:4: operator - not defined on: StringKind
