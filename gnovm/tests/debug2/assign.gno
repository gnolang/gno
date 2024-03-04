package main

func main() {
	x := "hello"
	x -= 1
	println(x)
}

// Error:
// main/debug/assign.gno:5: operator -= not defined on: StringKind
