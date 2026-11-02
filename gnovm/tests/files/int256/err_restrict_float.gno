package main

// The int256 restriction also applies to float constants.
const f float64 = 1 << 100

func main() {
	println(int256(f))
}

// Output:
// 1267650600228229401496703205376

// TypeCheckError:
// main/err_restrict_float.gno:7:17: cannot convert f (constant 1267650600228229401496703205376 of type float64) to type int256
