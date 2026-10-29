package main

func main() {
	var s uint = 32
	var x interface{} = ('a' << s) + (1 << 40)
	println(x)
}

// Error:
// main/shift_context5.gno:5:6-44: bigint overflows target kind

// TypeCheckError:
// main/shift_context5.gno:5:36: 1 << 40 (untyped rune constant 1099511627776) overflows rune; main/shift_context5.gno:5:35: (1 << 40) (untyped rune constant 1099511627776) overflows rune
