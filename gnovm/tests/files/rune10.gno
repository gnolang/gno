package main

// Converting an untyped rune constant to int32 checks the exact value.

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	var x int32 = c
	println(x)
}

// Error:
// main/rune10.gno:8:6-17: bigint overflows target kind

// TypeCheckError:
// main/rune10.gno:8:16: cannot use c (untyped rune constant 1241243320321) as int32 value in variable declaration (overflows)
