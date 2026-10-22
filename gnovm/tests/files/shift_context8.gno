package main

func main() {
	var s uint = 32
	var r rune = (1 << s) + (1 << 40)
	println(r)
}

// Error:
// main/shift_context8.gno:5:6-35: bigint overflows target kind

// TypeCheckError:
// main/shift_context8.gno:5:27: 1 << 40 (untyped int constant 1099511627776) overflows rune; main/shift_context8.gno:5:26: (1 << 40) (untyped int constant 1099511627776) overflows rune
