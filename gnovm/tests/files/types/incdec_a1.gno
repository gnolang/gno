package main

func main() {
	a := "hello"
	a++
	println(a)
}

// Error:
// main/incdec_a1.gno:5:2-5: operator ++ not defined on: StringKind

// TypeCheckError:
// main/incdec_a1.gno:5:2: invalid operation: a++ (non-numeric type string)
