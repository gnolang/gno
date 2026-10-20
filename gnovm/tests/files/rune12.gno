package main

// An explicit rune conversion checks the exact value.

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	println(rune(c))
}

// Error:
// main/rune12.gno:8:10-17: bigint overflows target kind

// TypeCheckError:
// main/rune12.gno:8:15: constant 1241243320321 overflows rune
