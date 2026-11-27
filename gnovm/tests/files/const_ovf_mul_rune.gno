package main

func main() {
	const a rune = '\U0010ffff'
	println(a * a)
}

// Error:
// main/const_ovf_mul_rune.gno:5:10-15: a * a (constant 1241243320321 of type int32) overflows int32

// TypeCheckError:
// main/const_ovf_mul_rune.gno:5:10: a * a (constant 1241243320321 of type rune) overflows rune
