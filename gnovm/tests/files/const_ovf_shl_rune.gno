package main

func main() {
	const a rune = -1
	const b = a << 32
	println(b)
}

// Error:
// main/const_ovf_shl_rune.gno:5:12-19: a << 32 (constant -4294967296 of type int32) overflows int32

// TypeCheckError:
// main/const_ovf_shl_rune.gno:5:12: a << 32 (constant -4294967296 of type rune) overflows rune
