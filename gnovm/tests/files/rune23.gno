package main

// As rune20, with an explicit conversion in the expression.

const N = 40

const C = 'a' << uint(N)

func main() {
	println(C)
}

// Error:
// main/rune23.gno:7:11-25: untyped rune constant 'a' << uint(N) overflows int32 (value 106652627894272); convert an operand to a wider integer type, e.g. int64('a')

// TypeCheckError:
// main/rune23.gno:10:10: cannot use C (untyped rune constant 106652627894272) as rune value in argument to built-in println (overflows)
