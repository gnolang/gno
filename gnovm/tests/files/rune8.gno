package main

// An untyped rune constant beyond int32 is rejected, even where gc accepts
// it: Gno's untyped runes hold int32 values. Use a typed operand (rune5).

func main() {
	var w int64 = '\U0010ffff' * '\U0010ffff'
	println(w)
}

// Error:
// main/rune8.gno:7:16-43: untyped rune constant '\U0010ffff' * '\U0010ffff' overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64('\U0010ffff')
