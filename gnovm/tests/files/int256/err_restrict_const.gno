package main

// An int256 constant must fit in an int64 (a language restriction).
const c int256 = 1 << 100

func main() {
	println(c)
}

// Output:
// 1267650600228229401496703205376

// TypeCheckError:
// main/err_restrict_const.gno:4:18: cannot use 1 << 100 (untyped int constant 1267650600228229401496703205376) as int256 value in constant declaration (overflows)
