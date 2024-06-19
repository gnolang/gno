package main

func main() {
	a := "hello"
	a++
	println(a)
}

// Error:
// main/files/types/incdec_a1.gno:5: operator ++ not defined on: StringKind
