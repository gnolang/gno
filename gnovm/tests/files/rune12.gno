package main

// As rune8, where the expression comes back within int32 (gc: the rune
// 1114111).

func main() {
	r := '\U0010ffff' * '\U0010ffff' / '\U0010ffff'
	println(r)
}

// Error:
// main/rune12.gno:7:7-34: untyped rune constant '\U0010ffff' * '\U0010ffff' overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64('\U0010ffff')
