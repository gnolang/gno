package main

const c int = 5

func main() {
	c++
	println("ok")
}

// Error:
// main/incdec_a5.gno:6:2-5: cannot assign to (const (5 int))

// TypeCheckError:
// main/incdec_a5.gno:6:2: cannot assign to c (neither addressable nor a map index expression)
