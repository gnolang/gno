package main

// Negating the minimum typed constant is not representable (#6392).

const g int32 = -2147483648

func main() {
	println(-g)
}

// Error:
// main/const_typed_overflow5.gno:8:10-12: -g (constant 2147483648 of type int32) overflows int32

// TypeCheckError:
// main/const_typed_overflow5.gno:8:10: -g (constant 2147483648 of type int32) overflows int32
