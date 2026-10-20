package main

const m int256 = -1 << 255

func main() {
	println(-m)
}

// Output:
// -57896044618658097711785492504343953926634992332820282019728792003956564819968

// TypeCheckError:
// main/err_fold_neg.gno:3:18: cannot use -1 << 255 (untyped int constant -57896044618658097711785492504343953926634992332820282019728792003956564819968) as int256 value in constant declaration (overflows)
