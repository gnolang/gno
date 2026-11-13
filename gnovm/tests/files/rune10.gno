package main

// As rune8, for a shift.

func main() {
	var x int64 = 'a' << 40
	println(x)
}

// Error:
// main/rune10.gno:5:1-8:2: untyped rune constant 'a' << 40 overflows int32 (value 106652627894272); convert an operand to a wider integer type, e.g. int64('a')
