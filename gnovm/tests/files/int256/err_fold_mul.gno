package main

const c int256 = 1 << 200

func main() {
	println(c * c)
}

// Output:
// 0

// TypeCheckError:
// main/err_fold_mul.gno:3:18: cannot use 1 << 200 (untyped int constant 1606938044258990275541962092341162602522202993782792835301376) as int256 value in constant declaration (overflows)
