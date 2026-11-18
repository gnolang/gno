package main

// As rune20, for imported constants.

import "unicode"

func main() {
	var x int64 = unicode.MaxRune * unicode.MaxRune
	println(x)
}

// Error:
// main/rune21.gno:8:16-49: untyped rune constant unicode.MaxRune * unicode.MaxRune overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64(unicode.MaxRune)
