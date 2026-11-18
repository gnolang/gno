package main

// A typed rune constant shifted past int32 is rejected, not zeroed
// (#6392).

const a rune = -1

func main() {
	println(a << 32)
}

// Error:
// main/const_typed_overflow8.gno:9:10-17: a << 32 (constant -4294967296 of type int32) overflows int32

// TypeCheckError:
// main/const_typed_overflow8.gno:9:10: a << 32 (constant -4294967296 of type rune) overflows rune
