package main

func main() {
	var r rune = 'a' + (1 << 40)
	println(r)
}

// Error:
// main/rune_arith_kind1.gno:4:6-30: bigint overflows target kind

// TypeCheckError:
// main/rune_arith_kind1.gno:4:15: cannot use 'a' + (1 << 40) (untyped rune constant 1099511627873) as rune value in variable declaration (overflows)
