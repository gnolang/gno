package main

// As rune20, for a shift by iota: D is 'Ā' << 24, beyond int32.

const (
	A = 'Ā' << (iota * 8)
	B
	C
	D
)

func main() {
	println(A, B, C, D)
}

// Error:
// main/rune22.gno:1:1-14:2: untyped rune constant 'Ā' << (iota * 8) overflows int32 (value 4294967296); convert an operand to a wider integer type, e.g. int64('Ā')

// TypeCheckError:
// main/rune22.gno:13:19: cannot use D (untyped rune constant 4294967296) as rune value in argument to built-in println (overflows)
