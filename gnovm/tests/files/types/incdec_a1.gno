package main

func main() {
	a := "hello"
	a++
	println(a)
}

// Error:
// main/files/types/incdec_a1.gno:5:2: operator ++ not defined on: StringKind

// TypeCheckError:
// main/files/types/incdec_a1.gno:5:2: invalid operation: a++ (non-numeric type string)
