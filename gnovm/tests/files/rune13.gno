package main

// As rune8, where the expression cancels to 0 (gc: rune 0).

func main() {
	r := '\U0010ffff' * '\U0010ffff' * '\x00'
	println(r)
}

// Error:
// main/rune13.gno:6:7-34: untyped rune constant '\U0010ffff' * '\U0010ffff' overflows int32 (value 1241243320321); convert an operand to a wider integer type, e.g. int64('\U0010ffff')
