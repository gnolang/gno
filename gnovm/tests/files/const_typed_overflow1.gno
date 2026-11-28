package main

// A typed integer constant operation must be representable in its type
// (#6392): the exact product, not the wrapped one.

const c int8 = 100

func main() {
	println(c * 2)
}

// Error:
// main/const_typed_overflow1.gno:9:10-15: c * 2 (constant 200 of type int8) overflows int8

// TypeCheckError:
// main/const_typed_overflow1.gno:9:10: c * 2 (constant 200 of type int8) overflows int8
