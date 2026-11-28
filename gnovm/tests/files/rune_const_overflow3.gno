package main

// A typed rune constant operation must be representable in rune (#6392).

const a rune = '\U0010ffff'

func main() {
	println(a * a)
}

// Error:
// main/rune_const_overflow3.gno:8:10-15: a * a (constant 1241243320321 of type int32) overflows int32

// TypeCheckError:
// main/rune_const_overflow3.gno:8:10: a * a (constant 1241243320321 of type rune) overflows rune
